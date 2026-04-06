package wuyinkeji

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

func (p *Provider) StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error) {
	if p.apiKey == "" {
		return "", errors.New("速创API 未配置 API Key")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = "video_veo3.1_fast"
	}
	if isSora2Model(model) {
		return p.startSora2Job(ctx, req)
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return "", errors.New("prompt is required")
	}
	startURL, err := p.buildStartURL(model)
	if err != nil {
		return "", err
	}
	payload := buildVideoPayload(model, req)
	raw, _ := json.Marshal(payload)
	logging.DownstreamRequest("provider_wuyin_video_start", p.ProviderName(), http.MethodPost, startURL, map[string]any{
		"model":  model,
		"prompt": logging.DownstreamPrompt(prompt),
	})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader(raw))
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", p.apiKey)
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponse("provider_wuyin_video_start_response", p.ProviderName(), http.MethodPost, startURL, 0, time.Since(start), err)
		return "", err
	}
	defer resp.Body.Close()
	b, _ := ioReadAllLimit(resp.Body, 4<<20)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponse("provider_wuyin_video_start_response", p.ProviderName(), http.MethodPost, startURL, resp.StatusCode, time.Since(start), errors.New("bad status"), string(b))
		return "", fmt.Errorf("wuyinkeji video start error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponse("provider_wuyin_video_start_response", p.ProviderName(), http.MethodPost, startURL, resp.StatusCode, time.Since(start), nil, string(b))
	var out startResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.Data.ID == "" {
		return "", fmt.Errorf("wuyinkeji video start missing id: %s", string(b))
	}
	return out.Data.ID, nil
}

func (p *Provider) GetVideoJob(ctx context.Context, jobID string) (string, string, string, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return "", "", "", errors.New("job_id is required")
	}
	u := fmt.Sprintf("%s/api/async/detail?key=%s&id=%s", p.baseURL, p.apiKey, jobID)
	logging.DownstreamRequestDebug("provider_wuyin_video_detail", p.ProviderName(), http.MethodGet, u, map[string]any{"job_id": jobID})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	hreq.Header.Set("Authorization", p.apiKey)
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseDebug("provider_wuyin_video_detail_response", p.ProviderName(), http.MethodGet, u, 0, time.Since(start), err)
		return "", "", "", err
	}
	defer resp.Body.Close()
	b, _ := ioReadAllLimit(resp.Body, 10<<20)
	logging.DownstreamResponseDebug("provider_wuyin_video_detail_response", p.ProviderName(), http.MethodGet, u, resp.StatusCode, time.Since(start), nil, string(b))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", "", fmt.Errorf("wuyinkeji video detail error: status=%d body=%s", resp.StatusCode, string(b))
	}
	var out detailResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", "", "", err
	}
	statusCode, msg := extractStatus(out.Data)
	status := mapVideoStatus(statusCode)
	urls := extractURLs(out.Data)
	videoURL := ""
	for _, u := range urls {
		if isVideoURL(u) {
			videoURL = u
			break
		}
	}
	return status, strings.TrimSpace(videoURL), strings.TrimSpace(msg), nil
}

func buildVideoPayload(model string, req types.VideoJobCreateRequest) map[string]any {
	payload := map[string]any{"prompt": strings.TrimSpace(req.Prompt)}
	if s := strings.TrimSpace(req.AspectRatio); s != "" {
		if strings.Contains(model, "grok") {
			payload["aspect_ratio"] = s
		} else {
			payload["aspectRatio"] = s
		}
	}
	if s := strings.TrimSpace(req.ImageSize); s != "" {
		payload["size"] = mapVideoSize(s)
	}
	if strings.Contains(model, "grok") && req.DurationSeconds > 0 {
		payload["duration"] = strconv.Itoa(req.DurationSeconds)
	}
	appendVideoRefs(payload, req)
	for k, v := range req.Extra {
		k = strings.TrimSpace(k)
		if k != "" && !types.IsCoreVideoField(k) && v != nil {
			payload[k] = v
		}
	}
	return payload
}

func appendVideoRefs(payload map[string]any, req types.VideoJobCreateRequest) {
	if s := strings.TrimSpace(req.Image); s != "" && isHTTPURL(s) {
		payload["firstFrameUrl"] = s
	}
	if s := extraString(req.Extra, "reference_url"); isHTTPURL(s) {
		payload["image_urls"] = []string{s}
	}
	if s := extraString(req.Extra, "last_frame_url"); isHTTPURL(s) {
		payload["lastFrameUrl"] = s
	}
}

func mapVideoStatus(status int) string {
	switch status {
	case 2:
		return "succeeded"
	case 3:
		return "failed"
	case 0:
		return "queued"
	default:
		return "running"
	}
}

func mapVideoSize(size string) string {
	switch strings.TrimSpace(strings.ToLower(size)) {
	case "1920x1080", "1080p":
		return "1080p"
	case "3840x2160", "4k":
		return "4K"
	default:
		return "720p"
	}
}

func isVideoURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") && (strings.Contains(s, ".mp4") || strings.Contains(s, ".mov") || strings.Contains(s, "video"))
}

func isHTTPURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
