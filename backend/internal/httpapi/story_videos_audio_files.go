package httpapi

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoMaterializeAudio(ctx context.Context, dir string, index int, resp types.AudioGenerateResponse) (string, error) {
	audioURL := strings.TrimSpace(resp.AudioURL)
	if audioURL == "" {
		return "", fmt.Errorf("未返回音频地址")
	}
	if strings.HasPrefix(audioURL, "/static/generated/") {
		src := filepath.Join(h.staticRoot, "generated", filepath.Base(audioURL))
		dst := filepath.Join(dir, fmt.Sprintf("segment-%02d%s", index, audioExt(src, resp.ContentType)))
		if err := copyLocalFile(src, dst); err != nil {
			return "", err
		}
		_ = os.Remove(src)
		return dst, nil
	}
	if strings.HasPrefix(audioURL, "http://") || strings.HasPrefix(audioURL, "https://") {
		return downloadAudioFile(ctx, dir, index, audioURL)
	}
	return "", fmt.Errorf("unsupported audio url: %s", audioURL)
}

func downloadAudioFile(ctx context.Context, dir string, index int, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download audio failed: status=%d", resp.StatusCode)
	}
	dst := filepath.Join(dir, fmt.Sprintf("segment-%02d%s", index, audioExt(rawURL, resp.Header.Get("Content-Type"))))
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(resp.Body, 200<<20))
	return dst, err
}

func copyLocalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func audioExt(pathOrURL, contentType string) string {
	if u, err := url.Parse(pathOrURL); err == nil {
		if ext := strings.TrimSpace(filepath.Ext(u.Path)); ext != "" {
			return ext
		}
	}
	if ext := strings.TrimSpace(filepath.Ext(pathOrURL)); ext != "" {
		return ext
	}
	if exts, _ := mime.ExtensionsByType(strings.TrimSpace(contentType)); len(exts) > 0 {
		return exts[0]
	}
	return ".audio"
}
