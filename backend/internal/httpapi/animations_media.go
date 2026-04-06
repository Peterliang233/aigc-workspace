package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"aigc-backend/internal/types"
	"aigc-backend/internal/util/mediafetch"
)

func (h *Handler) buildLeadImage(ctx context.Context, jobID string, job *types.AnimationJobCreateRequest) (string, error) {
	lead := strings.TrimSpace(job.LeadImage)
	if lead != "" {
		return h.normalizeAnimationImage(ctx, jobID, lead)
	}
	prov, ok := h.getImageProvider(job.Provider)
	if !ok || prov == nil {
		return "", fmt.Errorf("当前 provider 不支持自动生成首帧图，请手动上传参考图片")
	}
	req := types.ImageGenerateRequest{
		Provider:    job.Provider,
		Model:       "",
		Prompt:      job.Prompt,
		AspectRatio: job.AspectRatio,
		N:           1,
		Seed:        job.Seed,
	}
	if h.models != nil {
		req.Model = h.models.DefaultModel(job.Provider, "image")
	}
	h.applyImageModelDefaults(job.Provider, strings.TrimSpace(req.Model), &req)
	resp, err := prov.GenerateImage(ctx, req)
	if err != nil || len(resp.ImageURLs) == 0 {
		if err == nil {
			err = fmt.Errorf("自动首帧生成未返回图片")
		}
		return "", err
	}
	return h.normalizeAnimationImage(ctx, jobID, resp.ImageURLs[0])
}

func (h *Handler) normalizeAnimationImage(ctx context.Context, jobID, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	switch {
	case ref == "":
		return "", fmt.Errorf("empty image ref")
	case strings.HasPrefix(ref, "data:image/"):
		return ref, nil
	case strings.HasPrefix(ref, "/static/generated/"):
		return fileToDataURL(filepath.Join(h.staticRoot, "generated", filepath.Base(ref)), "")
	case strings.HasPrefix(ref, "/"):
		return "", fmt.Errorf("不支持内部相对路径参考图，请使用上传图片或公开 URL")
	default:
		return remoteImageToDataURL(ctx, jobID, ref)
	}
}

func remoteImageToDataURL(ctx context.Context, jobID, rawURL string) (string, error) {
	dir := filepath.Join(os.TempDir(), jobID+"_lead")
	dl := &mediafetch.Downloader{}
	path, ct, err := dl.DownloadToDirAutoExt(ctx, rawURL, dir, "lead", 20<<20)
	if err != nil {
		return "", err
	}
	return fileToDataURL(path, ct)
}

func fileToDataURL(path, contentType string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	ct := strings.TrimSpace(contentType)
	if ct == "" {
		ct = http.DetectContentType(b)
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

func localVideoToPath(ctx context.Context, dir, name, rawURL string) (string, error) {
	dl := &mediafetch.Downloader{}
	path, _, err := dl.DownloadToDirAutoExt(ctx, rawURL, dir, name, 300<<20)
	return path, err
}

func readCloserBytes(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	return io.ReadAll(r)
}
