package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"aigc-backend/internal/store"
)

const animationSegmentMaxAttempts = 3

var animationSegmentRetryBase = 2 * time.Second

func (h *Handler) runAnimationSegmentWithRetry(
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
	var localVideoPath string
	var nextLead string
	err := retryAnimationSegment(ctx, animationSegmentMaxAttempts, func(callCtx context.Context, attempt int) error {
		videoPath, lead, runErr := h.runAnimationSegmentOnce(callCtx, jobID, index, duration, segmentPrompt, leadImage, tmpDir, vp, job)
		if runErr == nil {
			localVideoPath = videoPath
			nextLead = lead
			return nil
		}
		if attempt < animationSegmentMaxAttempts && shouldRetryAnimationSegmentErr(runErr) {
			wait := time.Duration(attempt) * animationSegmentRetryBase
			slog.Default().Warn("animation_segment_retry_scheduled",
				"job_id", jobID,
				"segment", index+1,
				"attempt", attempt,
				"max_attempts", animationSegmentMaxAttempts,
				"wait_ms", wait.Milliseconds(),
				"err", runErr.Error(),
			)
			h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
				j.Status = "running"
				j.CurrentSegment = index + 1
				j.Segments[index].Status = "retrying"
				j.Segments[index].Error = retryMessage(attempt, runErr)
				j.Segments[index].SourceJobID = ""
				j.Segments[index].VideoURL = ""
				j.Segments[index].LastFramePath = ""
			})
			return runErr
		}
		h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
			j.Segments[index].Status = "failed"
			j.Segments[index].Error = runErr.Error()
		})
		return runErr
	})
	if err != nil {
		return "", "", err
	}
	return localVideoPath, nextLead, nil
}

func retryAnimationSegment(ctx context.Context, maxAttempts int, fn func(context.Context, int) error) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := fn(ctx, attempt)
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt == maxAttempts || !shouldRetryAnimationSegmentErr(err) {
			break
		}
		wait := time.Duration(attempt) * animationSegmentRetryBase
		if !sleepWithContext(ctx, wait) {
			if cerr := ctx.Err(); cerr != nil {
				return cerr
			}
			break
		}
	}
	return lastErr
}

func shouldRetryAnimationSegmentErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	for _, marker := range []string{
		"media worker 未配置",
		"provider 不可用",
		"empty image ref",
		"不支持自动生成首帧图",
	} {
		if strings.Contains(msg, strings.ToLower(marker)) {
			return false
		}
	}
	return true
}

func retryMessage(attempt int, err error) string {
	return fmt.Sprintf("第 %d 次尝试失败，准备重试：%s", attempt, err.Error())
}
