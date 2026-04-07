package gptbest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

func (p *Provider) GenerateAudio(ctx context.Context, req types.AudioGenerateRequest) (types.AudioGenerateResponse, error) {
	if p.apiKey == "" || p.baseURL == "" {
		return types.AudioGenerateResponse{}, errors.New("平台未配置 Base URL 或 API Key")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.imageModel
	}
	if model == "" {
		model = "gpt-4o-mini-tts"
	}
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return types.AudioGenerateResponse{}, errors.New("input is required")
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "alloy"
	}
	format := strings.ToLower(strings.TrimSpace(req.ResponseFormat))
	if format == "" {
		format = "mp3"
	}

	body := map[string]any{
		"model":           model,
		"input":           input,
		"voice":           voice,
		"response_format": format,
	}
	if req.Speed != nil {
		body["speed"] = *req.Speed
	}
	raw, _ := json.Marshal(body)
	u := p.baseURL + "/v1/audio/speech"
	logging.DownstreamRequestRaw("provider_gptbest_audio_speech", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":           model,
		"voice":           voice,
		"response_format": format,
		"speed":           req.Speed,
		"input":           logging.DownstreamPrompt(input),
	}, "application/json", raw)

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.doWithRetry(hreq, 3)
	if err != nil {
		logging.DownstreamResponseRaw("provider_gptbest_audio_speech_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err, "", nil)
		return types.AudioGenerateResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		logging.DownstreamResponseRaw("provider_gptbest_audio_speech_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		return types.AudioGenerateResponse{}, errors.New(strings.TrimSpace(string(b)))
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return types.AudioGenerateResponse{}, err
	}
	logging.DownstreamResponseRaw("provider_gptbest_audio_speech_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)
	ct := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if ct == "" || strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "application/octet-stream") {
		ct = contentTypeByAudioFormat(format)
	}

	ext := extByAudioFormat(format)
	if exts, _ := mime.ExtensionsByType(ct); len(exts) > 0 && strings.TrimSpace(exts[0]) != "" {
		ext = exts[0]
	}
	dir := filepath.Join(p.staticRoot, "generated")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return types.AudioGenerateResponse{}, err
	}
	f, err := os.CreateTemp(dir, "audio-*"+ext)
	if err != nil {
		return types.AudioGenerateResponse{}, err
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return types.AudioGenerateResponse{}, err
	}

	return types.AudioGenerateResponse{
		AudioURL:    "/static/generated/" + filepath.Base(f.Name()),
		Provider:    p.ProviderName(),
		Model:       model,
		ContentType: ct,
	}, nil
}

func contentTypeByAudioFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "flac":
		return "audio/flac"
	case "aac":
		return "audio/aac"
	case "opus":
		return "audio/opus"
	default:
		return "audio/mpeg"
	}
}

func extByAudioFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "flac":
		return ".flac"
	case "aac":
		return ".aac"
	case "opus":
		return ".opus"
	default:
		return ".mp3"
	}
}
