package httpapi

import (
	"context"
	"strings"
	"time"

	anim "aigc-backend/internal/animation"
	"aigc-backend/internal/store"
)

func (h *Handler) buildAnimationSegments(ctx context.Context, jobID string, job *store.AnimationJob, plan []int) []anim.PlannedSegment {
	fallback := anim.BuildFallbackSegments(job.Prompt, plan)
	if h.animationPlan == nil || !h.animationPlan.Enabled() {
		h.setAnimationPlan(jobID, "disabled", "", "", fallback)
		return fallback
	}
	plannerModel := strings.TrimSpace(job.PlannerModel)
	if plannerModel == "" {
		plannerModel = h.animationPlan.Model()
	}
	h.setAnimationPlan(jobID, "running", plannerModel, "", nil)
	timeoutCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	segments, err := h.animationPlan.Plan(timeoutCtx, plannerModel, anim.PromptPlanRequest{
		Prompt:           job.Prompt,
		TotalSeconds:     job.DurationSeconds,
		SegmentDurations: plan,
	})
	if err != nil {
		h.setAnimationPlan(jobID, "fallback", plannerModel, err.Error(), fallback)
		return fallback
	}
	h.setAnimationPlan(jobID, "succeeded", plannerModel, "", segments)
	return segments
}

func (h *Handler) setAnimationPlan(jobID, status, model, plannerErr string, segments []anim.PlannedSegment) {
	h.animationJobs.Update(jobID, func(j *store.AnimationJob) {
		j.PlannerStatus = status
		j.PlannerModel = model
		j.PlannerError = plannerErr
		for idx := range segments {
			if idx >= len(j.Segments) {
				break
			}
			j.Segments[idx].Prompt = segments[idx].Prompt
			j.Segments[idx].Continuity = segments[idx].Continuity
		}
	})
}
