package gptbest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

func (p *Provider) StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error) {
	if p.apiKey == "" || p.baseURL == "" {
		return "", errors.New("平台未配置 Base URL 或 API Key")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		return "", errors.New("model is required")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return "", errors.New("prompt is required")
	}
	body := req.MergePayload(map[string]any{
		"model":        model,
		"prompt":       prompt,
		"aspect_ratio": pickAspectRatio(req.AspectRatio, req.ImageSize),
		"duration":     pickVideoDuration(model, req.DurationSeconds),
	})
	if s := strings.TrimSpace(req.NegativePrompt); s != "" {
		body["negative_prompt"] = s
	}
	if req.Seed != nil {
		body["seed"] = *req.Seed
	}
	if img := strings.TrimSpace(req.Image); img != "" {
		applyGptBestVideoImage(body, img)
	}

	raw, _ := json.Marshal(body)
	cands := []string{p.baseURL + "/v2/videos/generations", p.baseURL + "/v1/video/generations"}
	lastErr := errors.New("视频任务创建失败")
	for i, u := range cands {
		id, retryNext, err := p.startVideoOnce(ctx, raw, body, u)
		if err == nil {
			return id, nil
		}
		lastErr = err
		if !retryNext || i == len(cands)-1 {
			break
		}
	}
	return "", lastErr
}

func (p *Provider) startVideoOnce(
	ctx context.Context,
	raw []byte,
	body map[string]any,
	u string,
) (string, bool, error) {
	logging.DownstreamRequest("provider_gptbest_video_start", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":           body["model"],
		"duration":        body["duration"],
		"aspect_ratio":    body["aspect_ratio"],
		"negative_prompt": logging.DownstreamPrompt(fmt.Sprint(body["negative_prompt"])),
		"prompt":          logging.DownstreamPrompt(fmt.Sprint(body["prompt"])),
		"image": func() any {
			s := firstNonEmptyString(body["img_url"], body["image"])
			if s == "" {
				return ""
			}
			if strings.HasPrefix(strings.ToLower(s), "http://") || strings.HasPrefix(strings.ToLower(s), "https://") {
				return logging.RedactURL(s)
			}
			return "base64"
		}(),
	})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.doWithRetry(hreq, 2)
	if err != nil {
		logging.DownstreamResponse("provider_gptbest_video_start_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err)
		return "", false, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponse("provider_gptbest_video_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"))
		retryNext := resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed
		return "", retryNext, fmt.Errorf("%s video start API error: status=%d body=%s", p.ProviderName(), resp.StatusCode, string(b))
	}
	logging.DownstreamResponse("provider_gptbest_video_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil)

	jobID := pickJSONStr(b, "task_id", "id", "job_id", "request_id")
	if jobID == "" {
		return "", false, errors.New("video start returned empty task_id")
	}
	return jobID, false, nil
}

func (p *Provider) GetVideoJob(ctx context.Context, jobID string) (string, string, string, error) {
	if p.apiKey == "" || p.baseURL == "" {
		return "", "", "", errors.New("平台未配置 Base URL 或 API Key")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return "", "", "", errors.New("job_id is required")
	}
	cands := []string{
		p.baseURL + "/v2/videos/generations/" + jobID,
		p.baseURL + "/v1/video/generations/" + jobID,
	}
	lastErr := errors.New("查询任务状态失败")
	for i, u := range cands {
		s, v, e, retryNext, err := p.getVideoOnce(ctx, u)
		if err == nil {
			return s, v, e, nil
		}
		lastErr = err
		if !retryNext || i == len(cands)-1 {
			break
		}
	}
	return "", "", "", lastErr
}

func (p *Provider) getVideoOnce(ctx context.Context, u string) (string, string, string, bool, error) {
	logging.DownstreamRequestDebug("provider_gptbest_video_status", p.ProviderName(), http.MethodGet, u, nil)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseDebug("provider_gptbest_video_status_response", p.ProviderName(), http.MethodGet, u, 0, time.Since(start), err)
		return "", "", "", false, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseDebug("provider_gptbest_video_status_response", p.ProviderName(), http.MethodGet, u, resp.StatusCode, time.Since(start), errors.New("bad status"))
		retryNext := resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed
		return "", "", "", retryNext, fmt.Errorf("%s video status API error: status=%d body=%s", p.ProviderName(), resp.StatusCode, string(b))
	}
	logging.DownstreamResponseDebug("provider_gptbest_video_status_response", p.ProviderName(), http.MethodGet, u, resp.StatusCode, time.Since(start), nil)
	status, jobErr := parseVideoStatusAndError(b)
	videoURL := pickVideoURL(b)
	return status, strings.TrimSpace(videoURL), strings.TrimSpace(jobErr), false, nil
}

func pickDuration(v int) int {
	if v > 0 {
		return v
	}
	return 5
}

func pickVideoDuration(model string, v int) any {
	if strings.EqualFold(strings.TrimSpace(model), "sora-2") {
		return fmt.Sprintf("%d", pickDuration(v))
	}
	return pickDuration(v)
}

func pickAspectRatio(ar, imageSize string) string {
	if s := strings.TrimSpace(ar); s != "" {
		return s
	}
	switch strings.TrimSpace(imageSize) {
	case "720x1280":
		return "9:16"
	case "960x960":
		return "1:1"
	default:
		return "16:9"
	}
}
