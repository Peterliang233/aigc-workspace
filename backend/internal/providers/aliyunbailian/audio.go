package aliyunbailian

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

type audioResponse struct {
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Output     struct {
		Audio struct {
			URL string `json:"url"`
		} `json:"audio"`
	} `json:"output"`
}

func (p *Provider) GenerateAudio(ctx context.Context, req types.AudioGenerateRequest) (types.AudioGenerateResponse, error) {
	if p.apiKey == "" || p.baseURL == "" {
		return types.AudioGenerateResponse{}, errors.New("阿里云百炼未配置 Base URL 或 API Key")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = "qwen3-tts-flash"
	}
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return types.AudioGenerateResponse{}, errors.New("input is required")
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "Cherry"
	}
	languageType := strings.TrimSpace(req.LanguageType)
	if languageType == "" {
		languageType = "Auto"
	}

	payloadInput := map[string]any{
		"text":          input,
		"voice":         voice,
		"language_type": languageType,
	}
	if instructions := strings.TrimSpace(req.Instructions); instructions != "" {
		payloadInput["instructions"] = instructions
	}
	body := map[string]any{
		"model": model,
		"input": payloadInput,
	}
	raw, _ := json.Marshal(body)
	u := p.generationURL()
	logging.DownstreamRequestRaw("provider_aliyun_bailian_audio", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":         model,
		"voice":         voice,
		"language_type": languageType,
		"instructions":  logging.DownstreamPrompt(req.Instructions),
		"text":          logging.DownstreamPrompt(input),
	}, "application/json", raw)

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseRaw("provider_aliyun_bailian_audio_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err, "", nil)
		return types.AudioGenerateResponse{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseRaw("provider_aliyun_bailian_audio_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		return types.AudioGenerateResponse{}, fmt.Errorf("aliyun bailian audio error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponseRaw("provider_aliyun_bailian_audio_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)

	url, err := parseAudioURL(b)
	if err != nil {
		return types.AudioGenerateResponse{}, err
	}
	return types.AudioGenerateResponse{
		AudioURL:    url,
		Provider:    p.ProviderName(),
		Model:       model,
		ContentType: "audio/wav",
	}, nil
}

func parseAudioURL(raw []byte) (string, error) {
	var out audioResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.StatusCode >= 400 || strings.TrimSpace(out.Code) != "" {
		msg := strings.TrimSpace(out.Message)
		if msg == "" {
			msg = strings.TrimSpace(out.Code)
		}
		return "", fmt.Errorf("aliyun bailian audio failed: %s", msg)
	}
	url := strings.TrimSpace(out.Output.Audio.URL)
	if url == "" {
		return "", fmt.Errorf("aliyun bailian audio missing url: %s", string(raw))
	}
	return url, nil
}
