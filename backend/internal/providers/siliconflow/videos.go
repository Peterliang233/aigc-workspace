package siliconflow

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

// Implements:
// - https://docs.siliconflow.cn/cn/api-reference/videos/videos_submit
// - https://docs.siliconflow.cn/cn/api-reference/videos/get_videos_status
//
// SiliconFlow is async: submit returns requestId, status returns output URL when Succeed.
type VideoProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewVideo(baseURL, apiKey string) *VideoProvider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.siliconflow.cn"
	}
	return &VideoProvider{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (p *VideoProvider) ProviderName() string { return "siliconflow" }

func (p *VideoProvider) StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error) {
	if p.apiKey == "" {
		return "", errors.New("SiliconFlow 未配置 API Key")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		return "", errors.New("model is required")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return "", errors.New("prompt is required")
	}

	imageSize := strings.TrimSpace(req.ImageSize)
	if imageSize == "" {
		imageSize = "1280x720"
	}

	image := strings.TrimSpace(req.Image)
	if strings.Contains(strings.ToUpper(model), "I2V") && image == "" {
		return "", errors.New("image is required for I2V models")
	}

	body := sfVideoSubmitRequest{
		Model:          model,
		Prompt:         prompt,
		NegativePrompt: strings.TrimSpace(req.NegativePrompt),
		ImageSize:      imageSize,
		Image:          image,
		Seed:           req.Seed,
	}
	raw, _ := json.Marshal(body)

	u := p.baseURL + "/v1/video/submit"
	logging.DownstreamRequest("provider_siliconflow_video_submit", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":           model,
		"image_size":      body.ImageSize,
		"seed_set":        body.Seed != nil,
		"negative_prompt": logging.DownstreamPrompt(body.NegativePrompt),
		"prompt":          logging.DownstreamPrompt(prompt),
		"image": func() any {
			if body.Image == "" {
				return ""
			}
			if strings.HasPrefix(strings.ToLower(body.Image), "http://") || strings.HasPrefix(strings.ToLower(body.Image), "https://") {
				return logging.RedactURL(body.Image)
			}
			return "base64"
		}(),
	})

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponse("provider_siliconflow_video_submit_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err)
		return "", err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponse("provider_siliconflow_video_submit_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"))
		return "", fmt.Errorf("siliconflow video submit API error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponse("provider_siliconflow_video_submit_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil)

	var out sfVideoSubmitResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if strings.TrimSpace(out.RequestID) == "" {
		return "", errors.New("siliconflow submit returned empty requestId")
	}
	return strings.TrimSpace(out.RequestID), nil
}

func (p *VideoProvider) GetVideoJob(ctx context.Context, jobID string) (string, string, string, error) {
	if p.apiKey == "" {
		return "", "", "", errors.New("SiliconFlow 未配置 API Key")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return "", "", "", errors.New("job_id is required")
	}

	raw, _ := json.Marshal(sfVideoStatusRequest{RequestID: jobID})
	u := p.baseURL + "/v1/video/status"
	logging.DownstreamRequestDebug("provider_siliconflow_video_status", p.ProviderName(), http.MethodPost, u, map[string]any{
		"requestId": jobID,
	})

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseDebug("provider_siliconflow_video_status_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err)
		return "", "", "", err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseDebug("provider_siliconflow_video_status_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"))
		return "", "", "", fmt.Errorf("siliconflow video status API error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponseDebug("provider_siliconflow_video_status_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil)

	var out sfVideoStatusResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return "", "", "", err
	}

	status := strings.TrimSpace(out.Status)
	switch status {
	case "Succeed":
		status = "succeeded"
	case "Failed":
		status = "failed"
	case "InQueue":
		status = "queued"
	case "InProgress":
		status = "running"
	default:
		status = "running"
	}

	videoURL := ""
	if len(out.Results.Videos) > 0 {
		videoURL = strings.TrimSpace(out.Results.Videos[0].URL)
	}
	return status, videoURL, strings.TrimSpace(out.Reason), nil
}
