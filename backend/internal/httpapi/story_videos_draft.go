package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/storyvideo"
	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	var req types.StoryVideoDraftRequest
	if err := decodeJSONWithLimit(w, r, &req, 2<<20); err != nil {
		return
	}
	req = h.storyVideoNormalizeDraftRequest(req)
	if len(req.Keywords) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "至少提供一个关键词"})
		return
	}
	planner, providerID, ok := h.storyPlannerFor(req.PlannerProvider)
	if !ok || planner == nil || !planner.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "故事脚本规划模型未配置: " + providerID})
		return
	}
	req.PlannerProvider = providerID
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	storyReq := storyvideo.DraftRequest{
		Keywords:        req.Keywords,
		Theme:           req.Theme,
		Audience:        req.Audience,
		Tone:            req.Tone,
		DurationSeconds: req.DurationSeconds,
		AspectRatio:     req.AspectRatio,
		Extra:           req.Extra,
	}
	var draft storyvideo.Draft
	err := retryDownstreamCall(ctx, "story_video_planner", func(callCtx context.Context) error {
		var callErr error
		draft, callErr = planner.Draft(callCtx, req.PlannerModel, storyReq)
		return callErr
	})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	resp, err := h.storyVideoCreateDraft(ctx, req, draft)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) storyVideoDraftID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	if projectID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var req types.StoryVideoDraftUpdateRequest
	if err := decodeJSONWithLimit(w, r, &req, 2<<20); err != nil {
		return
	}
	resp, err := h.storyVideoReplaceDraft(r.Context(), projectID, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) storyVideoNormalizeDraftRequest(req types.StoryVideoDraftRequest) types.StoryVideoDraftRequest {
	req.Keywords = storyVideoCleanList(req.Keywords)
	req.Theme = strings.TrimSpace(req.Theme)
	req.Audience = strings.TrimSpace(req.Audience)
	req.Tone = strings.TrimSpace(req.Tone)
	req.Extra = strings.TrimSpace(req.Extra)
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = 30
	}
	if req.DurationSeconds > 180 {
		req.DurationSeconds = 180
	}
	if strings.TrimSpace(req.AspectRatio) == "" {
		req.AspectRatio = "16:9"
	}
	req.PlannerProvider = strings.ToLower(strings.TrimSpace(req.PlannerProvider))
	if req.PlannerProvider == "" {
		req.PlannerProvider = h.defaultStoryPlannerProviderID()
	}
	if strings.TrimSpace(req.PlannerModel) == "" {
		req.PlannerModel = h.defaultStoryPlannerModel(req.PlannerProvider)
	}
	if strings.TrimSpace(req.ImageProvider) == "" && h.models != nil {
		req.ImageProvider = h.models.DefaultProvider("image")
	}
	if strings.TrimSpace(req.AudioProvider) == "" {
		req.AudioProvider = h.defaultAudioProviderID()
	}
	if strings.TrimSpace(req.ImageModel) == "" && h.models != nil {
		req.ImageModel = h.models.DefaultModel(req.ImageProvider, "image")
	}
	if strings.TrimSpace(req.AudioModel) == "" && h.models != nil {
		req.AudioModel = h.models.DefaultModel(req.AudioProvider, "audio")
	}
	return req
}

func storyVideoCleanList(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func storyVideoKeywordsJSON(items []string) string {
	b, _ := json.Marshal(storyVideoCleanList(items))
	return string(b)
}
