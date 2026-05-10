package httpapi

import (
	"aigc-backend/internal/types"
	"context"
	"net/http"
	"strings"
)

func (h *Handler) storyVideoConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	project, shots, err := h.storyVideos.GetProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if strings.TrimSpace(project.Status) != "draft_ready" {
		writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
		return
	}
	_ = h.storyVideos.UpdateProject(r.Context(), projectID, map[string]any{"status": "draft_confirmed", "error": nil, "video_asset_id": nil})
	_ = h.storyVideoAddEvent(r.Context(), projectID, "pipeline", "confirmed", "已确认开始生成素材", nil)
	go h.runStoryVideoPipeline(projectID)
	project.Status = "draft_confirmed"
	project.VideoAssetID = nil
	project.Error = nil
	writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
}

func (h *Handler) storyVideoRegenerateAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	var req types.StoryVideoRegenerateAudioRequest
	if r.ContentLength > 0 {
		if err := decodeJSONWithLimit(w, r, &req, 1<<20); err != nil {
			return
		}
	}
	if err := h.storyVideoPrepareAudioRetry(r.Context(), projectID, req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	go h.runStoryVideoAudio(projectID)
	project, shots, _ := h.storyVideos.GetProject(r.Context(), projectID)
	writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
}

func (h *Handler) storyVideoRegenerateShotImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	shotID := storyVideoShotID(r.URL.Path)
	var req types.StoryVideoRegenerateShotRequest
	if r.ContentLength > 0 {
		if err := decodeJSONWithLimit(w, r, &req, 1<<20); err != nil {
			return
		}
	}
	if err := h.storyVideoPrepareShotRetry(r.Context(), projectID, shotID, req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	go h.runStoryVideoShot(projectID, shotID)
	project, shots, _ := h.storyVideos.GetProject(r.Context(), projectID)
	writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
}

func (h *Handler) runStoryVideoPipeline(projectID string) {
	if h.storyVideos == nil {
		return
	}
	ctx := context.Background()
	_ = h.storyVideos.UpdateProject(ctx, projectID, map[string]any{"status": "images_generating", "error": nil, "video_asset_id": nil})
	_ = h.storyVideoAddEvent(ctx, projectID, "images", "started", "开始生成分镜图片", nil)
	project, shots, err := h.storyVideos.GetProject(ctx, projectID)
	if err != nil {
		return
	}
	for _, shot := range shots {
		if err := h.storyVideoGenerateShot(ctx, project, &shot); err != nil {
			_ = h.storyVideoAddEvent(ctx, projectID, "images", "failed", err.Error(), map[string]any{"shot_id": shot.ID})
		}
	}
	_ = h.storyVideoAddEvent(ctx, projectID, "audio", "started", "开始逐分镜生成解说音频，并在生成后拼接为一条完整音频", nil)
	_ = h.storyVideos.UpdateProject(ctx, projectID, map[string]any{"status": "audio_generating"})
	_ = h.runStoryVideoAudioSync(ctx, projectID)
	h.storyVideoRefreshProject(ctx, projectID)
}

func (h *Handler) runStoryVideoAudio(projectID string) {
	_ = h.runStoryVideoAudioSync(context.Background(), projectID)
	h.storyVideoRefreshProject(context.Background(), projectID)
}

func (h *Handler) runStoryVideoShot(projectID, shotID string) {
	project, _, err := h.storyVideos.GetProject(context.Background(), projectID)
	if err != nil {
		return
	}
	shot, err := h.storyVideos.GetShot(context.Background(), shotID)
	if err != nil {
		return
	}
	_ = h.storyVideoGenerateShot(context.Background(), project, shot)
	h.storyVideoRefreshProject(context.Background(), projectID)
}
