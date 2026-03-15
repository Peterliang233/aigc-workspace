package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

func (h *Handler) videosJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.videoProv == nil {
		slog.Default().Warn("videos_create_disabled")
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "视频能力暂未启用"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	var req types.VideoJobCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	slog.Default().Info("videos_create", "provider", h.videoProv.ProviderName())
	jobID, err := h.videoProv.StartVideoJob(ctx, req)
	if err != nil {
		slog.Default().Warn("videos_create_failed", "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	now := time.Now()
	h.jobs.Put(&store.VideoJob{
		ID:              jobID,
		Status:          "queued",
		Prompt:          req.Prompt,
		DurationSeconds: req.DurationSeconds,
		AspectRatio:     req.AspectRatio,
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	writeJSON(w, http.StatusOK, types.VideoJobCreateResponse{
		JobID:    jobID,
		Status:   "queued",
		Provider: h.videoProv.ProviderName(),
	})
}

func (h *Handler) videosJobsID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.videoProv == nil {
		slog.Default().Warn("videos_get_disabled")
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "视频能力暂未启用"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/videos/jobs/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing job id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	status, videoURL, jobErr, err := h.videoProv.GetVideoJob(ctx, id)
	if err != nil {
		slog.Default().Warn("videos_get_failed", "job_id", id, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	slog.Default().Info("videos_get", "job_id", id, "status", status)

	_ = h.jobs.Update(id, func(j *store.VideoJob) {
		j.Status = status
		j.VideoURL = videoURL
		j.Error = jobErr
		j.UpdatedAt = time.Now()
	})

	// If the job succeeded, persist video into MinIO and rewrite URL to this backend.
	if h.assets != nil && h.assets.Enabled() && status == "succeeded" && strings.TrimSpace(videoURL) != "" {
		job, _ := h.jobs.Get(id)
		prompt := ""
		duration := 0
		aspect := ""
		if job != nil {
			prompt = job.Prompt
			duration = job.DurationSeconds
			aspect = job.AspectRatio
		}
		a, err := h.assets.StoreRemote(ctx, assets.StoreRemoteInput{
			Capability:    "video",
			Provider:      strings.ToLower(strings.TrimSpace(h.videoProv.ProviderName())),
			Model:         strings.TrimSpace(h.effectiveCfg().VideoModel),
			Prompt:        prompt,
			ExternalJobID: id,
			Params: map[string]any{
				"duration_seconds": duration,
				"aspect_ratio":     aspect,
			},
			SourceURL: videoURL,
		})
		if err != nil {
			slog.Default().Warn("videos_store_asset_failed", "job_id", id, "err", err.Error())
		} else {
			videoURL = fmt.Sprintf("/api/assets/%d", a.ID)
			_ = h.jobs.Update(id, func(j *store.VideoJob) {
				j.VideoURL = videoURL
				j.UpdatedAt = time.Now()
			})
		}
	}

	writeJSON(w, http.StatusOK, types.VideoJobGetResponse{
		JobID:    id,
		Status:   status,
		VideoURL: videoURL,
		Error:    jobErr,
		Provider: h.videoProv.ProviderName(),
	})
}

