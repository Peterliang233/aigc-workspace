package httpapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	anim "aigc-backend/internal/animation"
	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

func (h *Handler) runAnimationJob(jobID string) {
	job, ok := h.animationJobs.Get(jobID)
	if !ok || job == nil {
		return
	}
	tmpDir, err := os.MkdirTemp("", jobID+"_")
	if err != nil {
		h.failAnimationJob(jobID, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	plan, _, err := anim.BuildPlan(job.DurationSeconds, h.animationDurationOptions(job.Provider, job.Model))
	if err != nil {
		h.failAnimationJob(jobID, err)
		return
	}
	h.prepareAnimationJob(jobID, plan)
	segments := h.buildAnimationSegments(context.Background(), jobID, job, plan)
	req := types.AnimationJobCreateRequest{
		Provider:        job.Provider,
		Model:           job.Model,
		Prompt:          job.Prompt,
		DurationSeconds: job.DurationSeconds,
		AspectRatio:     job.AspectRatio,
		LeadImage:       job.LeadImage,
		Seed:            job.Seed,
	}
	leadImage, err := h.buildLeadImage(context.Background(), jobID, &req)
	if err != nil {
		h.failAnimationJob(jobID, err)
		return
	}
	vp, ok := h.getVideoProvider(job.Provider)
	if !ok || vp == nil {
		h.failAnimationJob(jobID, fmt.Errorf("动画视频 provider 不可用"))
		return
	}

	localVideos := make([]string, 0, len(plan))
	for idx, dur := range plan {
		segmentPrompt := segments[idx].Prompt
		videoURL, nextLead, err := h.runAnimationSegment(context.Background(), jobID, idx, dur, segmentPrompt, leadImage, tmpDir, vp, job)
		if err != nil {
			h.failAnimationJob(jobID, err)
			return
		}
		localVideos = append(localVideos, videoURL)
		leadImage = nextLead
	}
	h.markAnimationStitching(jobID)

	outputPath := filepath.Join(tmpDir, "animation-final.mp4")
	if h.mediaWorker == nil || !h.mediaWorker.Enabled() {
		h.failAnimationJob(jobID, fmt.Errorf("media worker 未配置"))
		return
	}
	if err := h.mediaWorker.ConcatAndTrim(context.Background(), localVideos, outputPath, job.DurationSeconds); err != nil {
		h.failAnimationJob(jobID, err)
		return
	}
	url, err := h.storeAnimationFinal(context.Background(), outputPath, job)
	if err != nil {
		h.failAnimationJob(jobID, err)
		return
	}
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
		j.Status = "succeeded"
		j.VideoURL = url
		j.CurrentSegment = 0
	})
}

func (h *Handler) runAnimationSegment(
	ctx context.Context,
	jobID string,
	index int,
	duration int,
	segmentPrompt string,
	leadImage string,
	tmpDir string,
	vp videoProvider,
	job *store.AnimationJob,
) (string, string, error) {
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
		j.Status = "running"
		j.CurrentSegment = index + 1
		j.Segments[index].Status = "running"
		j.Segments[index].Prompt = segmentPrompt
	})
	vreq := types.VideoJobCreateRequest{
		Provider:        job.Provider,
		Model:           job.Model,
		Prompt:          segmentPrompt,
		DurationSeconds: duration,
		AspectRatio:     job.AspectRatio,
		Image:           leadImage,
		Seed:            job.Seed,
	}
	h.applyVideoModelDefaults(job.Provider, job.Model, &vreq)
	sourceJobID, err := vp.StartVideoJob(ctx, vreq)
	if err != nil {
		return "", "", err
	}
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) { j.Segments[index].SourceJobID = sourceJobID })
	remoteURL, err := h.pollAnimationVideo(ctx, jobID, index, sourceJobID, vp)
	if err != nil {
		return "", "", err
	}
	localVideoPath, err := localVideoToPath(ctx, tmpDir, fmt.Sprintf("segment-%02d", index+1), remoteURL)
	if err != nil {
		return "", "", err
	}
	framePath := filepath.Join(tmpDir, fmt.Sprintf("segment-%02d-last.jpg", index+1))
	if h.mediaWorker == nil || !h.mediaWorker.Enabled() {
		return "", "", fmt.Errorf("media worker 未配置")
	}
	if err := h.mediaWorker.ExtractLastFrame(ctx, localVideoPath, framePath); err != nil {
		return "", "", err
	}
	nextLead, err := fileToDataURL(framePath, "image/jpeg")
	if err != nil {
		return "", "", err
	}
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
		j.CompletedSegments++
		j.Segments[index].Status = "succeeded"
		j.Segments[index].VideoURL = remoteURL
		j.Segments[index].LastFramePath = framePath
	})
	return localVideoPath, nextLead, nil
}

func (h *Handler) pollAnimationVideo(ctx context.Context, jobID string, index int, sourceJobID string, vp videoProvider) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 45*time.Minute)
	defer cancel()
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for {
		status, videoURL, jobErr, err := vp.GetVideoJob(timeoutCtx, sourceJobID)
		if err != nil {
			return "", err
		}
		h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
			j.Segments[index].Status = status
			j.Segments[index].VideoURL = strings.TrimSpace(videoURL)
			j.Segments[index].Error = strings.TrimSpace(jobErr)
		})
		if status == "failed" {
			if jobErr == "" {
				jobErr = "分段生成失败"
			}
			return "", fmt.Errorf(jobErr)
		}
		if status == "succeeded" && strings.TrimSpace(videoURL) != "" {
			return strings.TrimSpace(videoURL), nil
		}
		select {
		case <-timeoutCtx.Done():
			return "", timeoutCtx.Err()
		case <-ticker.C:
		}
	}
}
