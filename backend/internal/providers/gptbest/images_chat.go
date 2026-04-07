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

type Provider struct {
	providerID string
	baseURL    string
	apiKey     string
	imageModel string
	staticRoot string
	httpClient *http.Client
}

func New(providerID, baseURL, apiKey, imageModel, staticRoot string) *Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.bltcy.ai"
	}
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		providerID = "bltcy"
	}
	return &Provider{
		providerID: providerID,
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		imageModel: strings.TrimSpace(imageModel),
		staticRoot: staticRoot,
		httpClient: newHTTPClient(),
	}
}

func (p *Provider) ProviderName() string { return p.providerID }

type chatReq struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *Provider) GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error) {
	if p.apiKey == "" || p.baseURL == "" {
		return types.ImageGenerateResponse{}, errors.New("平台未配置 Base URL 或 API Key")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return types.ImageGenerateResponse{}, errors.New("prompt is required")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.imageModel
	}
	if model == "" {
		model = "gpt-4o-image"
	}
	refs := mergeImageRefs(req.Image, req.ReferenceURLs)

	if len(refs) > 0 {
		out, err := p.generateByEdits(ctx, model, prompt, req, refs)
		if err == nil {
			return out, nil
		}
		if isEditOnlyModel(model) {
			return types.ImageGenerateResponse{}, err
		}
	}
	return p.generateByChat(ctx, model, prompt, refs)
}

func (p *Provider) generateByChat(ctx context.Context, model, prompt string, refs []string) (types.ImageGenerateResponse, error) {
	body := chatReq{Model: model, Messages: []chatMessage{{Role: "user", Content: buildChatUserContent(prompt, refs)}}}
	raw, _ := json.Marshal(body)
	u := chatCompletionsURL(p.baseURL)

	logging.DownstreamRequestRaw("provider_gptbest_request", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":       model,
		"image_count": len(refs),
		"prompt":      logging.DownstreamPrompt(prompt),
	}, "application/json", raw)

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.doWithRetry(hreq, 3)
	if err != nil {
		logging.DownstreamResponseRaw("provider_gptbest_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err, "", nil)
		if isTLSHandshakeTimeout(err) {
			return types.ImageGenerateResponse{}, errors.New("连接 BLTCY 失败：TLS 握手超时，请检查网络或稍后重试")
		}
		return types.ImageGenerateResponse{}, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 12<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseRaw("provider_gptbest_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"), resp.Header.Get("Content-Type"), b)
		return types.ImageGenerateResponse{}, fmt.Errorf("%s API error: status=%d body=%s", p.ProviderName(), resp.StatusCode, string(b))
	}
	logging.DownstreamResponseRaw("provider_gptbest_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), b)

	var out chatResp
	if err := json.Unmarshal(b, &out); err != nil {
		return types.ImageGenerateResponse{}, err
	}

	ref := extractImageRef(out)
	if ref == "" {
		return types.ImageGenerateResponse{}, errors.New("provider returned no image url")
	}
	if strings.HasPrefix(strings.ToLower(ref), "data:image/") {
		var err error
		ref, err = p.storeDataURL(ref)
		if err != nil {
			return types.ImageGenerateResponse{}, err
		}
	}

	return types.ImageGenerateResponse{
		ImageURLs: []string{ref},
		Provider:  p.ProviderName(),
		Model:     model,
	}, nil
}

func buildChatUserContent(prompt string, refs []string) any {
	if len(refs) == 0 {
		return prompt
	}
	content := []map[string]any{{"type": "text", "text": prompt}}
	for _, ref := range refs {
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": ref},
		})
	}
	return content
}

func isEditOnlyModel(model string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(model)), "qwen-image-edit")
}

func chatCompletionsURL(base string) string {
	u := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(strings.ToLower(u), "/v1/chat/completions") {
		return u
	}
	return u + "/v1/chat/completions"
}
