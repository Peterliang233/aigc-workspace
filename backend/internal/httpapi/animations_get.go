package httpapi

import (
	"net/http"
	"strings"

	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

func (h *Handler) animationsJobsID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/animations/jobs/"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing job id"})
		return
	}
	job, ok := h.animationJobs.Get(id)
	if !ok || job == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "job not found"})
		return
	}
	writeJSON(w, http.StatusOK, animationJobResponse(job))
}

func animationJobResponse(job *store.AnimationJob) types.AnimationJobGetResponse {
	out := types.AnimationJobGetResponse{
		JobID:             job.ID,
		Status:            job.Status,
		Provider:          job.Provider,
		Model:             job.Model,
		Prompt:            job.Prompt,
		DurationSeconds:   job.DurationSeconds,
		PlannerStatus:     job.PlannerStatus,
		PlannerModel:      strings.TrimSpace(job.PlannerModel),
		PlannerError:      job.PlannerError,
		SegmentCount:      job.SegmentCount,
		CompletedSegments: job.CompletedSegments,
		CurrentSegment:    job.CurrentSegment,
		VideoURL:          job.VideoURL,
		Error:             job.Error,
	}
	if len(job.Segments) == 0 {
		return out
	}
	out.Segments = make([]types.AnimationSegmentResponse, 0, len(job.Segments))
	for _, seg := range job.Segments {
		out.Segments = append(out.Segments, types.AnimationSegmentResponse{
			Index:          seg.Index,
			Status:         seg.Status,
			Duration:       seg.Duration,
			Prompt:         strings.TrimSpace(seg.Prompt),
			Continuity:     strings.TrimSpace(seg.Continuity),
			SourceJobID:    seg.SourceJobID,
			VideoURL:       seg.VideoURL,
			LastFrameReady: strings.TrimSpace(seg.LastFramePath) != "",
			Error:          seg.Error,
		})
	}
	return out
}
