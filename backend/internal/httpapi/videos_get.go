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

func (h *Handler) videosJobsID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

	job, _ := h.jobs.Get(id)
	providerID := ""
	if job != nil {
		providerID = strings.ToLower(strings.TrimSpace(job.Provider))
	}
	if providerID == "" {
		providerID = h.defaultVideoProviderID()
	}
	vp, ok := h.getVideoProvider(providerID)
	if !ok || vp == nil {
		slog.Default().Warn("videos_get_disabled", "job_id", id, "provider", providerID)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "视频能力暂未启用或未配置"})
		return
	}

	status, videoURL, jobErr, err := vp.GetVideoJob(ctx, id)
	if err != nil {
		slog.Default().Warn("videos_get_failed", "job_id", id, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	slog.Default().Info("videos_get", "job_id", id, "provider", providerID, "provider_impl", vp.ProviderName(), "status", status)

	h.jobs.Update(id, func(j *store.VideoJob) {
		j.Status = status
		j.VideoURL = videoURL
		j.Error = jobErr
		j.UpdatedAt = time.Now()
	})

	// If the job succeeded, persist video into MinIO and rewrite URL to this backend.
	if h.assets != nil && h.assets.Enabled() && status == "succeeded" && strings.TrimSpace(videoURL) != "" {
		params := map[string]any{}
		in := assets.StoreRemoteInput{
			Capability:    "video",
			Provider:      providerID,
			Model:         "",
			Prompt:        "",
			ExternalJobID: id,
			Params:        params,
			SourceURL:     videoURL,
		}
		if job != nil {
			in.Provider = strings.ToLower(strings.TrimSpace(job.Provider))
			in.Model = strings.TrimSpace(job.Model)
			in.Prompt = job.Prompt
			if job.DurationSeconds > 0 {
				params["duration_seconds"] = job.DurationSeconds
			}
			if strings.TrimSpace(job.AspectRatio) != "" {
				params["aspect_ratio"] = strings.TrimSpace(job.AspectRatio)
			}
			if strings.TrimSpace(job.ImageSize) != "" {
				params["image_size"] = strings.TrimSpace(job.ImageSize)
			}
		}

		a, err := h.assets.StoreRemote(ctx, in)
		if err != nil {
			slog.Default().Warn("videos_store_asset_failed", "job_id", id, "err", err.Error())
		} else {
			videoURL = fmt.Sprintf("/api/assets/%d", a.ID)
			h.jobs.Update(id, func(j *store.VideoJob) {
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
		Provider: providerID,
	})
}
