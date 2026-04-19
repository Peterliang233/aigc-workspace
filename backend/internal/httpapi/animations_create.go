package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

func (h *Handler) animationsJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req types.AnimationJobCreateRequest
	if err := decodeJSONWithLimit(w, r, &req, 24<<20); err != nil {
		return
	}
	providerID := strings.ToLower(strings.TrimSpace(req.Provider))
	if providerID == "" && h.models != nil {
		providerID = strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("video")))
	}
	if providerID == "" {
		providerID = h.defaultVideoProviderID()
	}
	vp, ok := h.getVideoProvider(providerID)
	if !ok || vp == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "动画能力暂未启用或未配置"})
		return
	}
	if strings.TrimSpace(req.Model) == "" && h.models != nil {
		req.Model = h.models.DefaultModel(providerID, "video")
	}
	if strings.TrimSpace(req.Model) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "缺少 model"})
		return
	}
	if !h.modelSupportsInitImage(providerID, req.Model) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "动画工坊仅支持图生视频模型（需支持 image 输入）"})
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "缺少 prompt"})
		return
	}
	if req.DurationSeconds <= 0 || req.DurationSeconds > 180 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "duration_seconds 需要在 1 到 180 秒之间"})
		return
	}
	jobID := fmt.Sprintf("anim_%d", time.Now().UnixNano())
	now := time.Now()
	job := &store.AnimationJob{
		ID:              jobID,
		Status:          "queued",
		Provider:        providerID,
		Model:           strings.TrimSpace(req.Model),
		PlannerModel:    strings.TrimSpace(req.PlannerModel),
		Prompt:          req.Prompt,
		DurationSeconds: req.DurationSeconds,
		AspectRatio:     strings.TrimSpace(req.AspectRatio),
		LeadImage:       strings.TrimSpace(req.LeadImage),
		Seed:            req.Seed,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	h.animationJobs.Put(job)
	go h.runAnimationJob(jobID)
	writeJSON(w, http.StatusOK, types.AnimationJobCreateResponse{
		JobID:           jobID,
		Status:          "queued",
		Provider:        providerID,
		Model:           strings.TrimSpace(req.Model),
		DurationSeconds: req.DurationSeconds,
	})
}
