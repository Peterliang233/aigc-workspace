package gptbest

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

type imageGenerationsReq struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

func (p *Provider) generateByImageAPI(
	ctx context.Context,
	model, prompt string,
	req types.ImageGenerateRequest,
) (types.ImageGenerateResponse, error) {
	body := imageGenerationsReq{
		Model:          model,
		Prompt:         prompt,
		Size:           strings.TrimSpace(req.Size),
		N:              maxImageCount(req.N),
		ResponseFormat: "b64_json",
	}
	raw, _ := json.Marshal(body)
	u := imagesGenerationsURL(p.baseURL)

	logging.DownstreamRequestRaw("provider_gptbest_request", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":           model,
		"size":            body.Size,
		"n":               body.N,
		"response_format": body.ResponseFormat,
		"prompt":          logging.DownstreamPrompt(prompt),
	}, "application/json", raw)

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.doWithRetry(hreq, 1)
	if err != nil {
		logging.DownstreamResponseRaw("provider_gptbest_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err, "", nil)
		return types.ImageGenerateResponse{}, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 12<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseRaw("provider_gptbest_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		return types.ImageGenerateResponse{}, fmt.Errorf("%s API error: status=%d body=%s", p.ProviderName(), resp.StatusCode, string(b))
	}
	logging.DownstreamResponseRaw("provider_gptbest_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)

	urls, err := p.parseImageAPIResponse(b)
	if err != nil {
		return types.ImageGenerateResponse{}, err
	}
	return types.ImageGenerateResponse{ImageURLs: urls, Provider: p.ProviderName(), Model: model}, nil
}

func prefersImagesAPIModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "gpt-image-")
}

func imagesGenerationsURL(base string) string {
	u := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(strings.ToLower(u), "/v1/images/generations") {
		return u
	}
	return u + "/v1/images/generations"
}

func maxImageCount(n int) int {
	if n <= 0 {
		return 1
	}
	return n
}
