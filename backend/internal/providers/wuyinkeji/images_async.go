package wuyinkeji

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
	"aigc-backend/internal/util/mediafetch"
)

// Implements the "速创API-图片模型-图片生成" family.
// Example doc: https://api.wuyinkeji.com/doc/65 (NanoBanana2)
//
// Their image generation endpoints are async:
// - POST /api/async/<model_endpoint>?key=API_KEY  => {data:{id:"..."}}
// - GET  /api/async/detail?key=API_KEY&id=...     => {data:{status:2, ... result ...}}
//
// We wrap it as a "sync" call to match this app's /api/images/generate:
// start job -> poll -> extract image URLs -> download to /static/ (best effort).
type Provider struct {
	baseURL string
	apiKey  string

	// Optional allow-list of models shown in UI. If non-empty, requests must pick one of them.
	// For "dynamic endpoint" mode, each model is the path segment in:
	//   POST /api/async/{model}?key=...
	models []string

	// Legacy mode: model label -> endpoint path (starting with /).
	// If configured, we use this mapping instead of the dynamic endpoint pattern.
	endpoints map[string]string

	httpClient *http.Client
	staticRoot string
}

func New(baseURL, apiKey, staticRoot string, models []string, endpoints map[string]string) *Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.wuyinkeji.com"
	}

	var ms []string
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		// normalize common full paths to "segment" form
		m = strings.TrimPrefix(m, "/api/async/")
		m = strings.TrimPrefix(m, "api/async/")
		ms = append(ms, m)
	}

	ep := map[string]string{}
	for k, v := range endpoints {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		if !strings.HasPrefix(v, "/") {
			v = "/" + v
		}
		ep[k] = v
	}

	return &Provider{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		models:     ms,
		endpoints:  ep,
		httpClient: &http.Client{Timeout: 10 * time.Minute},
		staticRoot: staticRoot,
	}
}

func (p *Provider) ProviderName() string { return "wuyinkeji" }

func (p *Provider) Models() []string {
	// Prefer legacy mapping keys if present; otherwise return configured model segments.
	if len(p.endpoints) > 0 {
		var out []string
		for m := range p.endpoints {
			out = append(out, m)
		}
		sort.Strings(out)
		return out
	}
	out := make([]string, 0, len(p.models))
	out = append(out, p.models...)
	sort.Strings(out)
	return out
}

type startResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (p *Provider) GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error) {
	if p.apiKey == "" {
		return types.ImageGenerateResponse{}, errors.New("速创API 未配置 API Key")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return types.ImageGenerateResponse{}, errors.New("prompt is required")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		// Choose a stable default.
		if len(p.models) > 0 {
			model = p.models[0]
		} else if len(p.endpoints) > 0 {
			// deterministic default in legacy mode
			var keys []string
			for k := range p.endpoints {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			model = keys[0]
		} else {
			model = "image_nanoBanana2"
		}
	}

	startURL := ""
	if len(p.endpoints) > 0 {
		ep, ok := p.endpoints[model]
		if !ok {
			return types.ImageGenerateResponse{}, fmt.Errorf("unknown model %q", model)
		}
		startURL = fmt.Sprintf("%s%s?key=%s", p.baseURL, ep, p.apiKey)
	} else {
		if len(p.models) > 0 && !contains(p.models, model) {
			return types.ImageGenerateResponse{}, fmt.Errorf("unknown model %q", model)
		}
		seg, err := normalizeModelSegment(model)
		if err != nil {
			return types.ImageGenerateResponse{}, err
		}
		startURL = fmt.Sprintf("%s/api/async/%s?key=%s", p.baseURL, seg, p.apiKey)
	}

	payload := map[string]any{
		"prompt":      prompt,
		"size":        mapSize(req.Size),
		"aspectRatio": mapAspect(req.AspectRatio),
	}

	logging.DownstreamRequest("provider_wuyin_start", p.ProviderName(), http.MethodPost, startURL, map[string]any{
		"model":  model,
		"prompt": logging.DownstreamPrompt(prompt),
		"size":   payload["size"],
		"aspect": payload["aspectRatio"],
	})

	raw, _ := json.Marshal(payload)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader(raw))
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponse("provider_wuyin_start_response", p.ProviderName(), http.MethodPost, startURL, 0, time.Since(start), err)
		slog.Default().Warn("provider_wuyin_start_failed", "err", err.Error())
		return types.ImageGenerateResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioReadAllLimit(resp.Body, 4<<20)
		logging.DownstreamResponse("provider_wuyin_start_response", p.ProviderName(), http.MethodPost, startURL, resp.StatusCode, time.Since(start), errors.New("bad status"))
		slog.Default().Warn("provider_wuyin_start_bad_status", "status", resp.StatusCode)
		return types.ImageGenerateResponse{}, fmt.Errorf("wuyinkeji start error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponse("provider_wuyin_start_response", p.ProviderName(), http.MethodPost, startURL, resp.StatusCode, time.Since(start), nil)

	var out startResp
	b, _ := ioReadAllLimit(resp.Body, 10<<20)
	if err := json.Unmarshal(b, &out); err != nil {
		return types.ImageGenerateResponse{}, err
	}
	if out.Data.ID == "" {
		slog.Default().Warn("provider_wuyin_start_missing_id")
		return types.ImageGenerateResponse{}, fmt.Errorf("wuyinkeji start missing id: %s", string(b))
	}

	slog.Default().Info("provider_wuyin_job_started", "job_id", out.Data.ID, "model", model)
	urls, err := p.pollResult(ctx, out.Data.ID)
	if err != nil {
		slog.Default().Warn("provider_wuyin_poll_failed", "job_id", out.Data.ID, "err", err.Error())
		return types.ImageGenerateResponse{}, err
	}
	if len(urls) == 0 {
		return types.ImageGenerateResponse{}, errors.New("wuyinkeji returned no image urls")
	}

	outDir := filepath.Join(p.staticRoot, "generated")
	var localURLs []string
	dl := &mediafetch.Downloader{HTTP: p.httpClient}
	for i, u := range urls {
		prefix := fmt.Sprintf("wy_%d_%d", time.Now().UnixNano(), i)
		dst, _, err := dl.DownloadToDirAutoExt(ctx, u, outDir, prefix, 25<<20)
		if err != nil {
			slog.Default().Warn("provider_wuyin_download_failed", "err", err.Error())
			// fallback to remote
			localURLs = append(localURLs, u)
			continue
		}
		localURLs = append(localURLs, "/static/generated/"+filepath.Base(dst))
	}

	return types.ImageGenerateResponse{
		ImageURLs: localURLs,
		Provider:  p.ProviderName(),
		Model:     model,
	}, nil
}

type detailResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (p *Provider) pollResult(ctx context.Context, id string) ([]string, error) {
	deadline := time.Now().Add(10 * time.Minute)

	for time.Now().Before(deadline) {
		u := fmt.Sprintf("%s/api/async/detail?key=%s&id=%s", p.baseURL, p.apiKey, id)
		logging.DownstreamRequestDebug("provider_wuyin_detail", p.ProviderName(), http.MethodGet, u, map[string]any{
			"job_id": id,
		})
		hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		hreq.Header.Set("Authorization", p.apiKey)

		start := time.Now()
		resp, err := p.httpClient.Do(hreq)
		if err != nil {
			logging.DownstreamResponseDebug("provider_wuyin_detail_response", p.ProviderName(), http.MethodGet, u, 0, time.Since(start), err)
			return nil, err
		}
		b, _ := ioReadAllLimit(resp.Body, 10<<20)
		resp.Body.Close()
		logging.DownstreamResponseDebug("provider_wuyin_detail_response", p.ProviderName(), http.MethodGet, u, resp.StatusCode, time.Since(start), nil)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("wuyinkeji detail error: status=%d body=%s", resp.StatusCode, string(b))
		}

		var out detailResp
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, err
		}

		status, msg := extractStatus(out.Data)
		switch status {
		case 2:
			slog.Default().Info("provider_wuyin_job_succeeded", "job_id", id)
			return extractURLs(out.Data), nil
		case 3:
			if msg == "" {
				msg = "wuyinkeji job failed"
			}
			slog.Default().Warn("provider_wuyin_job_failed", "job_id", id, "msg", msg)
			return nil, errors.New(msg)
		default:
			// queued/running
			select {
			case <-time.After(1200 * time.Millisecond):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, errors.New("wuyinkeji timeout waiting for result")
}

func mapSize(size string) string {
	// doc accepts: 1K,2K,4K
	s := strings.TrimSpace(strings.ToLower(size))
	switch s {
	case "1024x1024", "1k":
		return "1K"
	case "2048x2048", "2k":
		return "2K"
	case "4096x4096", "4k":
		return "4K"
	default:
		return "1K"
	}
}

func mapAspect(ar string) string {
	ar = strings.TrimSpace(ar)
	if ar == "" {
		return "1:1"
	}
	// allow common ratios used in our app
	switch ar {
	case "1:1", "16:9", "9:16", "4:3", "3:4", "2:3", "3:2":
		return ar
	default:
		return "1:1"
	}
}

var urlRe = regexp.MustCompile(`https?://[^\s"']+`)

var modelSegRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func normalizeModelSegment(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("model is required")
	}
	s = strings.TrimPrefix(s, "/api/async/")
	s = strings.TrimPrefix(s, "api/async/")
	s = strings.TrimPrefix(s, "/")
	// Only allow a single safe path segment to avoid SSRF/path traversal.
	if strings.Contains(s, "/") {
		return "", errors.New("invalid model name")
	}
	if !modelSegRe.MatchString(s) {
		return "", errors.New("invalid model name")
	}
	return s, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func extractURLs(v any) []string {
	// Robust fallback: scan for http(s) strings in the JSON-encoded data.
	b, _ := json.Marshal(v)
	m := urlRe.FindAllString(string(b), -1)
	seen := map[string]bool{}
	var out []string
	for _, u := range m {
		u = strings.Trim(u, `\"`)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

func extractStatus(v any) (status int, message string) {
	// Expected fields from docs:
	// data.status: 0 queued, 1 running, 2 success, 3 failed
	// data.message: error message
	m, ok := v.(map[string]any)
	if !ok {
		return 0, ""
	}
	if raw, ok := m["status"]; ok {
		switch t := raw.(type) {
		case float64:
			status = int(t)
		case int:
			status = t
		case string:
			// ignore
		}
	}
	if raw, ok := m["message"]; ok {
		if s, ok := raw.(string); ok {
			message = s
		}
	}
	return status, message
}

func ioReadAllLimit(r io.Reader, n int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, n))
}
