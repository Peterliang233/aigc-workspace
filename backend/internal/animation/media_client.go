package animation

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

func (c *MediaClient) ExtractLastFrame(ctx context.Context, inputPath, outputPath string) error {
	return c.uploadOne(ctx, "/extract-last-frame", "file", inputPath, outputPath, nil)
}

func (c *MediaClient) ConcatAndTrim(ctx context.Context, inputs []string, outputPath string, durationSeconds int) error {
	fields := map[string]string{}
	if durationSeconds > 0 {
		fields["duration_seconds"] = itoa(durationSeconds)
	}
	return c.uploadMany(ctx, "/concat-videos", "files", inputs, outputPath, fields)
}
