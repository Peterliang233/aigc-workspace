package storyvideo

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type MediaClient struct {
	baseURL string
	http    *http.Client
}

func NewMediaClient(baseURL string) *MediaClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &MediaClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Minute},
	}
}

func (c *MediaClient) Enabled() bool {
	return c != nil && c.baseURL != ""
}

func (c *MediaClient) ComposeSlideshow(ctx context.Context, images []string, audioPath string, durations []int, aspectRatio, outputPath string) error {
	fields := map[string]string{
		"durations_ms_json": intsJSON(durations),
		"aspect_ratio":      strings.TrimSpace(aspectRatio),
	}
	return c.uploadCompose(ctx, "/compose-slideshow", images, audioPath, outputPath, fields)
}

func (c *MediaClient) ConcatAudios(ctx context.Context, audioPaths []string, outputPath string) ([]int, error) {
	return c.uploadAudioConcat(ctx, audioPaths, outputPath)
}
