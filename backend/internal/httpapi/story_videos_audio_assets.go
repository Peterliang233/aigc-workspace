package httpapi

import (
	"context"
	"strings"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoClearShotAudioAssets(ctx context.Context, segments []storyVideoAudioSegment) error {
	for _, segment := range segments {
		if strings.TrimSpace(segment.ShotID) == "" {
			continue
		}
		if err := h.storyVideos.UpdateShot(ctx, segment.ShotID, map[string]any{"audio_asset_id": nil}); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) storyVideoStoreShotAudio(ctx context.Context, segment storyVideoAudioSegment, path, providerID, model string, req types.AudioGenerateRequest) error {
	if h.assets == nil || !h.assets.Enabled() || strings.TrimSpace(segment.ShotID) == "" {
		return nil
	}
	asset, err := h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
		Capability: "audio",
		Provider:   providerID,
		Model:      model,
		Prompt:     segment.Text,
		Params: map[string]any{
			"shot_id":       segment.ShotID,
			"voice":         req.Voice,
			"instructions":  req.Instructions,
			"language_type": req.LanguageType,
			"segment":       true,
		},
		FilePath: path,
	})
	if err != nil {
		return err
	}
	return h.storyVideos.UpdateShot(ctx, segment.ShotID, map[string]any{"audio_asset_id": asset.ID})
}
