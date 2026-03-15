package wuyinkeji

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"aigc-backend/internal/logging"
)

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
			select {
			case <-time.After(1200 * time.Millisecond):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, errors.New("wuyinkeji timeout waiting for result")
}

