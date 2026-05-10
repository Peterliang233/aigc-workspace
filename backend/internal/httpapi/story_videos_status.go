package httpapi

import (
	"context"
	"fmt"
	"strings"

	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoRefreshProject(ctx context.Context, projectID string) {
	project, shots, err := h.storyVideos.GetProject(ctx, projectID)
	if err != nil {
		return
	}
	hasFailed, hasPending := false, false
	for _, shot := range shots {
		switch strings.TrimSpace(shot.Status) {
		case "failed":
			hasFailed = true
		case "queued", "generating", "regenerating":
			hasPending = true
		}
	}
	status := "ready_to_compose"
	switch {
	case project.VideoAssetID != nil:
		status = "succeeded"
	case hasPending:
		status = "images_generating"
	case hasFailed && project.AudioAssetID == nil:
		status = "needs_attention"
	case hasFailed:
		status = "images_partial_failed"
	case project.AudioAssetID == nil:
		status = "audio_failed"
	}
	attrs := map[string]any{"status": status}
	if status == "succeeded" || status == "ready_to_compose" || status == "images_partial_failed" || status == "images_generating" {
		attrs["error"] = nil
	}
	_ = h.storyVideos.UpdateProject(ctx, projectID, attrs)
}

func (h *Handler) storyVideoPrepareAudioRetry(ctx context.Context, projectID string, req types.StoryVideoRegenerateAudioRequest) error {
	if _, _, err := h.storyVideos.GetProject(ctx, projectID); err != nil {
		return err
	}
	attrs := map[string]any{"status": "audio_generating", "audio_asset_id": nil, "video_asset_id": nil, "error": nil}
	if strings.TrimSpace(req.NarrationText) != "" {
		attrs["narration_text"] = strings.TrimSpace(req.NarrationText)
	}
	if strings.TrimSpace(req.AudioProvider) != "" {
		attrs["audio_provider"] = strings.TrimSpace(req.AudioProvider)
	}
	if strings.TrimSpace(req.AudioModel) != "" {
		attrs["audio_model"] = strings.TrimSpace(req.AudioModel)
	}
	if strings.TrimSpace(req.AudioVoice) != "" {
		attrs["audio_voice"] = strings.TrimSpace(req.AudioVoice)
	}
	_ = h.storyVideoAddEvent(ctx, projectID, "audio", "retry_requested", "已请求重新生成音频", nil)
	return h.storyVideos.UpdateProject(ctx, projectID, attrs)
}

func (h *Handler) storyVideoPrepareShotRetry(ctx context.Context, projectID, shotID string, req types.StoryVideoRegenerateShotRequest) error {
	if _, _, err := h.storyVideos.GetProject(ctx, projectID); err != nil {
		return err
	}
	shot, err := h.storyVideos.GetShot(ctx, shotID)
	if err != nil || shot.ProjectID != projectID {
		return fmt.Errorf("分镜不存在")
	}
	attrs := map[string]any{"status": "regenerating", "image_asset_id": nil, "error": nil}
	if strings.TrimSpace(req.ImagePrompt) != "" {
		attrs["image_prompt"] = strings.TrimSpace(req.ImagePrompt)
	}
	if err := h.storyVideos.UpdateShot(ctx, shotID, attrs); err != nil {
		return err
	}
	projectAttrs := map[string]any{"status": "images_generating", "video_asset_id": nil, "error": nil}
	if strings.TrimSpace(req.ImageProvider) != "" {
		projectAttrs["image_provider"] = strings.TrimSpace(req.ImageProvider)
	}
	if strings.TrimSpace(req.ImageModel) != "" {
		projectAttrs["image_model"] = strings.TrimSpace(req.ImageModel)
	}
	_ = h.storyVideoAddEvent(ctx, projectID, "images", "retry_requested", "已请求重新生成分镜图片", map[string]any{"shot_id": shotID})
	return h.storyVideos.UpdateProject(ctx, projectID, projectAttrs)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
