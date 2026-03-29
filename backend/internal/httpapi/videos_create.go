package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

func (h *Handler) videosJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	var req types.VideoJobCreateRequest
	if err := decodeJSONWithLimit(w, r, &req, 16<<20); err != nil {
		return
	}

	providerID := strings.ToLower(strings.TrimSpace(req.Provider))
	if providerID == "" {
		if h.models != nil {
			providerID = strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("video")))
		}
		if providerID == "" {
			providerID = h.defaultVideoProviderID()
		}
	}
	vp, ok := h.getVideoProvider(providerID)
	if !ok || vp == nil {
		slog.Default().Warn("videos_create_disabled", "provider", providerID)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "视频能力暂未启用或未配置"})
		return
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		if h.models != nil {
			model = strings.TrimSpace(h.models.DefaultModel(providerID, "video"))
			req.Model = model
		}
	}
	if strings.TrimSpace(model) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "缺少 model"})
		return
	}

	h.applyVideoModelDefaults(providerID, model, &req)
	if miss := h.missingVideoRequiredFields(providerID, model, req); len(miss) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing required fields: " + strings.Join(miss, ", ")})
		return
	}

	slog.Default().Info("videos_create", "provider", providerID, "provider_impl", vp.ProviderName(), "model", model)
	jobID, err := vp.StartVideoJob(ctx, req)
	if err != nil {
		slog.Default().Warn("videos_create_failed", "provider", providerID, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	now := time.Now()
	h.jobs.Put(&store.VideoJob{
		ID:              jobID,
		Status:          "queued",
		Provider:        providerID,
		Model:           model,
		Prompt:          req.Prompt,
		DurationSeconds: req.DurationSeconds,
		AspectRatio:     req.AspectRatio,
		ImageSize:       req.ImageSize,
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	writeJSON(w, http.StatusOK, types.VideoJobCreateResponse{
		JobID:    jobID,
		Status:   "queued",
		Provider: providerID,
	})
}
