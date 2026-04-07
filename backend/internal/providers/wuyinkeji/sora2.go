package wuyinkeji

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

func isSora2Model(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "video_sora2")
}

func (p *Provider) startSora2Job(ctx context.Context, req types.VideoJobCreateRequest) (string, error) {
	payload := map[string]any{"prompt": strings.TrimSpace(req.Prompt)}
	if s := extraString(req.Extra, "url"); isHTTPURL(s) {
		payload["url"] = s
	}
	if s := strings.TrimSpace(req.AspectRatio); s != "" {
		payload["aspectRatio"] = s
	}
	if req.DurationSeconds > 0 {
		payload["duration"] = fmt.Sprintf("%d", req.DurationSeconds)
	}
	if s := extraString(req.Extra, "size"); s != "" {
		payload["size"] = s
	}
	if s := extraString(req.Extra, "remix_target_id"); s != "" {
		payload["remixTargetId"] = s
	}
	raw, _ := json.Marshal(payload)
	u := fmt.Sprintf("%s/api/async/video_sora2?key=%s", p.baseURL, p.apiKey)
	logging.DownstreamRequestRaw("provider_wuyin_sora2_start", p.ProviderName(), http.MethodPost, u, map[string]any{
		"prompt": logging.DownstreamPrompt(req.Prompt),
		"params": payload,
	}, "application/json", raw)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseRaw("provider_wuyin_sora2_start_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err, "", nil)
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioReadAllLimit(resp.Body, 4<<20)
	logging.DownstreamResponseRaw("provider_wuyin_sora2_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), body)
	jobID, msg, err := parseSora2Start(body)
	if err != nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return "", err
		}
		return "", fmt.Errorf("wuyinkeji sora2 start error: status=%d body=%s", resp.StatusCode, string(body))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && jobID != "" {
		return jobID, nil
	}
	if msg == "" {
		msg = string(body)
	}
	return "", fmt.Errorf("wuyinkeji sora2 start failed: %s", msg)
}

func parseSora2Start(raw []byte) (string, string, error) {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", "", err
	}
	msg := strings.TrimSpace(fmt.Sprint(body["msg"]))
	switch data := body["data"].(type) {
	case map[string]any:
		if id := cleanAnyString(data["id"]); id != "" {
			return id, msg, nil
		}
	case []any:
		for _, item := range data {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if id := cleanAnyString(m["id"]); id != "" {
				return id, msg, nil
			}
		}
	case nil:
	}
	return "", msg, nil
}

func cleanAnyString(value any) string {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" || strings.EqualFold(text, "null") {
		return ""
	}
	return text
}
