package httpapi

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"aigc-backend/internal/assets"
	"github.com/minio/minio-go/v7"
)

func (h *Handler) storyVideoCompose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	if h.storyMedia == nil || !h.storyMedia.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "媒体合成服务未配置"})
		return
	}
	if h.assets == nil || !h.assets.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "素材存储未配置"})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	project, shots, err := h.storyVideos.GetProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if project.VideoAssetID != nil {
		writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
		return
	}
	if project.AudioAssetID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请先生成解说音频"})
		return
	}
	for _, shot := range shots {
		if shot.ImageAssetID == nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "仍有分镜图片未生成完成"})
			return
		}
	}
	_ = h.storyVideos.UpdateProject(r.Context(), projectID, map[string]any{"status": "composing", "error": nil})
	_ = h.storyVideoAddEvent(r.Context(), projectID, "compose", "started", "开始合成故事视频", nil)
	go h.runStoryVideoCompose(projectID)
	project.Status = "composing"
	writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
}

func (h *Handler) runStoryVideoCompose(projectID string) {
	ctx := context.Background()
	project, shots, err := h.storyVideos.GetProject(ctx, projectID)
	if err != nil {
		return
	}
	tmpDir, err := os.MkdirTemp("", "story-video-compose-*")
	if err != nil {
		_ = h.storyVideoFailProject(ctx, projectID, "compose", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	audioPath, err := h.storyVideoDownloadAsset(ctx, *project.AudioAssetID, tmpDir, "narration")
	if err != nil {
		_ = h.storyVideoFailProject(ctx, projectID, "compose", err)
		return
	}
	images, durations := make([]string, 0, len(shots)), make([]int, 0, len(shots))
	for i, shot := range shots {
		path, dlErr := h.storyVideoDownloadAsset(ctx, *shot.ImageAssetID, tmpDir, fmt.Sprintf("shot-%02d", i+1))
		if dlErr != nil {
			_ = h.storyVideoFailProject(ctx, projectID, "compose", dlErr)
			return
		}
		images = append(images, path)
		durations = append(durations, shot.DurationMS)
	}
	outputPath := filepath.Join(tmpDir, "story-video.mp4")
	if err := h.storyMedia.ComposeSlideshow(ctx, images, audioPath, durations, project.AspectRatio, outputPath); err != nil {
		_ = h.storyVideoFailProject(ctx, projectID, "compose", err)
		return
	}
	asset, err := h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
		Capability: "video", Provider: "storyvideo", Model: "slideshow", Prompt: project.Title, FilePath: outputPath, ContentType: "video/mp4",
	})
	if err != nil {
		_ = h.storyVideoFailProject(ctx, projectID, "compose", err)
		return
	}
	_ = h.storyVideos.UpdateProject(ctx, projectID, map[string]any{"status": "succeeded", "video_asset_id": asset.ID, "error": nil})
	_ = h.storyVideoAddEvent(ctx, projectID, "compose", "succeeded", "故事视频合成成功", map[string]any{"asset_id": asset.ID})
}

func (h *Handler) storyVideoDownloadAsset(ctx context.Context, assetID uint64, dir, name string) (string, error) {
	if h.assets == nil || !h.assets.Enabled() || h.assets.MinIO == nil {
		return "", fmt.Errorf("素材存储未配置")
	}
	asset, err := h.assets.Store.Get(ctx, assetID)
	if err != nil {
		return "", err
	}
	ext := ".bin"
	if exts, _ := mime.ExtensionsByType(asset.ContentType); len(exts) > 0 {
		ext = exts[0]
	}
	path := filepath.Join(dir, name+ext)
	obj, err := h.assets.MinIO.Client.GetObject(ctx, h.assets.MinIO.Bucket, asset.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	defer obj.Close()
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	_, err = io.Copy(dst, obj)
	return path, err
}
