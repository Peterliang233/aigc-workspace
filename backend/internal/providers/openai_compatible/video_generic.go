package openai_compatible

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

// Video provider is intentionally "generic async" because vendors differ a lot.
// It expects:
// - POST {baseURL}{startEndpoint} => {"id":"..."} (or {"job_id":"..."})
// - GET  {baseURL}{statusEndpoint with {id}} => {"status":"queued|running|succeeded|failed","output_url":"...","error":"..."}
type VideoProvider struct {
	baseURL    string
	apiKey     string
	videoModel string
	startEP    string
	statusEP   string
	httpClient *http.Client
}

func NewVideoGeneric(baseURL, apiKey, videoModel, startEndpoint, statusEndpoint string) *VideoProvider {
	return &VideoProvider{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		videoModel: strings.TrimSpace(videoModel),
		startEP:    strings.TrimSpace(startEndpoint),
		statusEP:   strings.TrimSpace(statusEndpoint),
		httpClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (p *VideoProvider) ProviderName() string { return "openai_compatible_video_generic" }

type startResp struct {
	ID    string `json:"id"`
	JobID string `json:"job_id"`
}

func (p *VideoProvider) StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error) {
	if p.baseURL == "" || p.apiKey == "" {
		return "", errors.New("平台未配置 Base URL 或 API Key")
	}
	if p.startEP == "" || p.statusEP == "" {
		return "", errors.New("视频能力未配置接口")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return "", errors.New("prompt is required")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.videoModel
	}
	if model == "" {
		return "", errors.New("missing model")
	}

	payload := req.MergePayload(map[string]any{
		"model":            model,
		"prompt":           prompt,
		"duration_seconds": req.DurationSeconds,
		"aspect_ratio":     req.AspectRatio,
	})
	raw, _ := json.Marshal(payload)

	u := p.baseURL + p.startEP
	logging.DownstreamRequestRaw("provider_video_start", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":            model,
		"duration_seconds": req.DurationSeconds,
		"aspect_ratio":     req.AspectRatio,
		"prompt":           logging.DownstreamPrompt(prompt),
	}, "application/json", raw)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseRaw("provider_video_start_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err, "", nil)
		return "", err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseRaw("provider_video_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		slog.Default().Warn("provider_video_start_bad_status", "status", resp.StatusCode)
		return "", fmt.Errorf("video start API error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponseRaw("provider_video_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)
	var out startResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.ID != "" {
		slog.Default().Info("provider_video_started", "job_id", out.ID)
		return out.ID, nil
	}
	if out.JobID != "" {
		slog.Default().Info("provider_video_started", "job_id", out.JobID)
		return out.JobID, nil
	}
	return "", errors.New("video start API returned no id/job_id")
}

type statusResp struct {
	Status    string `json:"status"`
	OutputURL string `json:"output_url"`
	VideoURL  string `json:"video_url"`
	Error     string `json:"error"`
}

func (p *VideoProvider) GetVideoJob(ctx context.Context, jobID string) (string, string, string, error) {
	if p.baseURL == "" || p.apiKey == "" {
		return "", "", "", errors.New("平台未配置 Base URL 或 API Key")
	}
	if p.statusEP == "" {
		return "", "", "", errors.New("视频能力未配置接口")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return "", "", "", errors.New("job_id is required")
	}

	ep := strings.ReplaceAll(p.statusEP, "{id}", jobID)
	u := p.baseURL + ep
	logging.DownstreamRequestDebugRaw("provider_video_status", p.ProviderName(), http.MethodGet, u, map[string]any{
		"job_id": jobID,
	}, "", nil)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseDebugRaw("provider_video_status_response", p.ProviderName(), http.MethodGet, u, 0, time.Since(start), err, "", nil)
		return "", "", "", err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseDebugRaw("provider_video_status_response", p.ProviderName(), http.MethodGet, u, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		slog.Default().Warn("provider_video_status_bad_status", "status", resp.StatusCode)
		return "", "", "", fmt.Errorf("video status API error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponseDebugRaw("provider_video_status_response", p.ProviderName(), http.MethodGet, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)

	var out statusResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", "", "", err
	}
	videoURL := out.VideoURL
	if videoURL == "" {
		videoURL = out.OutputURL
	}
	status := strings.TrimSpace(out.Status)
	if status == "" {
		status = "running"
	}
	return status, videoURL, out.Error, nil
}
