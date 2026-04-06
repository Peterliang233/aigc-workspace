package wuyinkeji

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

type audioResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
}

func (p *Provider) GenerateAudio(ctx context.Context, req types.AudioGenerateRequest) (types.AudioGenerateResponse, error) {
	if p.apiKey == "" {
		return types.AudioGenerateResponse{}, errors.New("速创API 未配置 API Key")
	}
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return types.AudioGenerateResponse{}, errors.New("input is required")
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "female-shaonv"
	}
	body := map[string]any{"text": input, "voice_id": voice}
	if req.Speed != nil {
		body["speed"] = *req.Speed
	}
	raw, _ := json.Marshal(body)
	u := p.baseURL + "/api/voice/composite"
	logging.DownstreamRequest("provider_wuyin_audio_start", p.ProviderName(), http.MethodPost, u, map[string]any{
		"voice_id": voice,
		"speed":    req.Speed,
		"text":     logging.DownstreamPrompt(input),
	})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Content-Type", "application/json;charset=utf-8")
	hreq.Header.Set("Authorization", p.apiKey)
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponse("provider_wuyin_audio_start_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err)
		return types.AudioGenerateResponse{}, err
	}
	defer resp.Body.Close()
	b, _ := ioReadAllLimit(resp.Body, 4<<20)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponse("provider_wuyin_audio_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"), string(b))
		return types.AudioGenerateResponse{}, fmt.Errorf("wuyinkeji audio error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponse("provider_wuyin_audio_start_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, string(b))
	var out audioResp
	if err := json.Unmarshal(b, &out); err != nil {
		return types.AudioGenerateResponse{}, err
	}
	if strings.TrimSpace(out.Data.URL) == "" {
		return types.AudioGenerateResponse{}, fmt.Errorf("wuyinkeji audio missing url: %s", string(b))
	}
	return types.AudioGenerateResponse{
		AudioURL:    strings.TrimSpace(out.Data.URL),
		Provider:    p.ProviderName(),
		Model:       strings.TrimSpace(req.Model),
		ContentType: "audio/mpeg",
	}, nil
}
