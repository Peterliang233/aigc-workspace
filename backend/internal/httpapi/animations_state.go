package httpapi

import (
	"context"
	"fmt"
	"log/slog"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/store"
)

func (h *Handler) storeAnimationFinal(ctx context.Context, outputPath string, job *store.AnimationJob) (string, error) {
	if h.assets == nil || !h.assets.Enabled() {
		return "", fmt.Errorf("MinIO 未配置，无法保存动画成片")
	}
	a, err := h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
		Capability: "video",
		Provider:   job.Provider,
		Model:      job.Model,
		Prompt:     job.Prompt,
		Params: map[string]any{
			"animation":         true,
			"duration_seconds":  job.DurationSeconds,
			"segment_count":     job.SegmentCount,
			"continuous_camera": true,
		},
		FilePath:    outputPath,
		ContentType: "video/mp4",
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/api/assets/%d", a.ID), nil
}

func (h *Handler) prepareAnimationJob(jobID string, plan []int) {
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
		j.Status = "planning"
		j.PlannerError = ""
		j.SegmentCount = len(plan)
		j.CompletedSegments = 0
		j.CurrentSegment = 0
		j.Segments = make([]store.AnimationSegment, 0, len(plan))
		for idx, dur := range plan {
			j.Segments = append(j.Segments, store.AnimationSegment{
				Index:    idx,
				Status:   "queued",
				Duration: dur,
			})
		}
	})
}

func (h *Handler) markAnimationStitching(jobID string) {
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) { j.Status = "stitching" })
}

func (h *Handler) failAnimationJob(jobID string, err error) {
	slog.Default().Warn("animation_job_failed", "job_id", jobID, "err", err.Error())
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
		j.Status = "failed"
		j.Error = err.Error()
		j.CurrentSegment = 0
	})
}
