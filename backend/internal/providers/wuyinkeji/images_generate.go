package wuyinkeji

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
	"aigc-backend/internal/util/mediafetch"
)

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

	startURL, err := p.buildStartURL(model)
	if err != nil {
		return types.ImageGenerateResponse{}, err
	}

	payload := map[string]any{
		"prompt":      prompt,
		"size":        mapSize(req.Size),
		"aspectRatio": mapAspect(req.AspectRatio),
	}
	raw, _ := json.Marshal(payload)

	logging.DownstreamRequestRaw("provider_wuyin_start", p.ProviderName(), http.MethodPost, startURL, map[string]any{
		"model":  model,
		"prompt": logging.DownstreamPrompt(prompt),
		"size":   payload["size"],
		"aspect": payload["aspectRatio"],
	}, "application/json", raw)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader(raw))
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseRaw("provider_wuyin_start_response", p.ProviderName(), http.MethodPost, startURL, 0, time.Since(start), err, "", nil)
		slog.Default().Warn("provider_wuyin_start_failed", "err", err.Error())
		return types.ImageGenerateResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioReadAllLimit(resp.Body, 4<<20)
		logging.DownstreamResponseRaw("provider_wuyin_start_response", p.ProviderName(), http.MethodPost, startURL, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		slog.Default().Warn("provider_wuyin_start_bad_status", "status", resp.StatusCode)
		return types.ImageGenerateResponse{}, fmt.Errorf("wuyinkeji start error: status=%d body=%s", resp.StatusCode, string(b))
	}
	var out startResp
	b, _ := ioReadAllLimit(resp.Body, 10<<20)
	logging.DownstreamResponseRaw("provider_wuyin_start_response", p.ProviderName(), http.MethodPost, startURL, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)
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

func (p *Provider) buildStartURL(model string) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", errors.New("model is required")
	}

	if len(p.endpoints) > 0 {
		ep, ok := p.endpoints[model]
		if !ok {
			return "", fmt.Errorf("unknown model %q", model)
		}
		return fmt.Sprintf("%s%s?key=%s", p.baseURL, ep, p.apiKey), nil
	}

	if len(p.models) > 0 && !contains(p.models, model) {
		return "", fmt.Errorf("unknown model %q", model)
	}
	seg, err := normalizeModelSegment(model)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/api/async/%s?key=%s", p.baseURL, seg, p.apiKey), nil
}
