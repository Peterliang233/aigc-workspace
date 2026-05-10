package httpapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/storyvideo"
	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoGenerateShot(ctx context.Context, project *storyvideo.Project, shot *storyvideo.Shot) error {
	defaultProvider, defaultModel := "", ""
	if h.models != nil {
		defaultProvider = h.models.DefaultProvider("image")
	}
	providerID := firstNonEmpty(project.ImageProvider, defaultProvider)
	if h.models != nil {
		defaultModel = h.models.DefaultModel(providerID, "image")
	}
	model := firstNonEmpty(project.ImageModel, defaultModel)
	prov, ok := h.getImageProvider(providerID)
	if !ok || prov == nil {
		return h.storyVideoFailShot(ctx, project.ID, shot.ID, fmt.Errorf("图片 provider 不可用"))
	}
	nextAttempt := shot.AttemptCount + 1
	_ = h.storyVideos.UpdateShot(ctx, shot.ID, map[string]any{"status": "generating", "attempt_count": nextAttempt, "error": nil})
	req := types.ImageGenerateRequest{Provider: providerID, Model: model, Prompt: shot.ImagePrompt, AspectRatio: project.AspectRatio, N: 1}
	h.applyImageModelDefaults(providerID, model, &req)
	if miss := h.missingImageRequiredFields(providerID, model, req); len(miss) > 0 {
		return h.storyVideoFailShot(ctx, project.ID, shot.ID, fmt.Errorf("缺少图片参数: %s", strings.Join(miss, ", ")))
	}
	var resp types.ImageGenerateResponse
	err := retryDownstreamCall(ctx, "story_video_image_generate", func(callCtx context.Context) error {
		var callErr error
		resp, callErr = prov.GenerateImage(callCtx, req)
		return callErr
	})
	if err != nil || len(resp.ImageURLs) == 0 {
		if err == nil {
			err = fmt.Errorf("未返回图片地址")
		}
		return h.storyVideoFailShot(ctx, project.ID, shot.ID, err)
	}
	asset, err := h.storyVideoStoreImageAsset(ctx, providerID, model, shot.ImagePrompt, req, resp.ImageURLs[0])
	if err != nil {
		return h.storyVideoFailShot(ctx, project.ID, shot.ID, err)
	}
	_ = h.storyVideos.UpdateShot(ctx, shot.ID, map[string]any{"status": "succeeded", "image_asset_id": asset.ID, "error": nil})
	_ = h.storyVideoAddEvent(ctx, project.ID, "images", "shot_succeeded", "分镜图片生成成功", map[string]any{"shot_id": shot.ID, "asset_id": asset.ID})
	return nil
}

func (h *Handler) storyVideoStoreImageAsset(ctx context.Context, providerID, model, prompt string, req types.ImageGenerateRequest, src string) (*assets.Asset, error) {
	if h.assets == nil || !h.assets.Enabled() {
		return nil, fmt.Errorf("素材存储未配置")
	}
	if strings.HasPrefix(src, "/static/generated/") {
		path := filepath.Join(h.staticRoot, "generated", filepath.Base(src))
		a, err := h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
			Capability: "image", Provider: providerID, Model: model, Prompt: prompt,
			Params: map[string]any{"size": req.Size, "aspect_ratio": req.AspectRatio}, FilePath: path,
		})
		if err == nil {
			_ = os.Remove(path)
		}
		return a, err
	}
	return h.assets.StoreRemote(ctx, assets.StoreRemoteInput{
		Capability: "image", Provider: providerID, Model: model, Prompt: prompt,
		Params: map[string]any{"size": req.Size, "aspect_ratio": req.AspectRatio}, SourceURL: strings.TrimSpace(src),
	})
}

func (h *Handler) storyVideoFailShot(ctx context.Context, projectID, shotID string, err error) error {
	msg := strings.TrimSpace(err.Error())
	_ = h.storyVideos.UpdateShot(ctx, shotID, map[string]any{"status": "failed", "error": msg})
	_ = h.storyVideoAddEvent(ctx, projectID, "images", "shot_failed", msg, map[string]any{"shot_id": shotID})
	return err
}

func (h *Handler) storyVideoFailProject(ctx context.Context, projectID, stage string, err error) error {
	msg := strings.TrimSpace(err.Error())
	_ = h.storyVideos.UpdateProject(ctx, projectID, map[string]any{"status": stage + "_failed", "error": msg})
	_ = h.storyVideoAddEvent(ctx, projectID, stage, "failed", msg, nil)
	return err
}
