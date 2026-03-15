package openai_compatible

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/types"
)

type Provider struct {
	baseURL    string
	apiKey     string
	imageModel string

	httpClient *http.Client

	staticRoot string
}

func New(baseURL, apiKey, imageModel, staticRoot string) *Provider {
	return &Provider{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		imageModel: strings.TrimSpace(imageModel),
		httpClient: &http.Client{Timeout: 2 * time.Minute},
		staticRoot: staticRoot,
	}
}

func (p *Provider) ProviderName() string { return "openai_compatible" }

type imagesGenerateRequest struct {
	Model          string `json:"model,omitempty"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"` // "b64_json" or "url"
}

type imagesGenerateResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
}

func (p *Provider) GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error) {
	if p.baseURL == "" || p.apiKey == "" {
		return types.ImageGenerateResponse{}, errors.New("missing AIGC_BASE_URL or AIGC_API_KEY")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return types.ImageGenerateResponse{}, errors.New("prompt is required")
	}

	n := req.N
	if n <= 0 {
		n = 1
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "1024x1024"
	}

	body := imagesGenerateRequest{
		Model:          p.imageModel,
		Prompt:         prompt,
		Size:           size,
		N:              n,
		ResponseFormat: "b64_json",
	}
	raw, _ := json.Marshal(body)

	u := p.baseURL + "/v1/images/generations"
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		return types.ImageGenerateResponse{}, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return types.ImageGenerateResponse{}, fmt.Errorf("images API error: status=%d body=%s", resp.StatusCode, string(b))
	}

	var out imagesGenerateResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return types.ImageGenerateResponse{}, err
	}

	if len(out.Data) == 0 {
		return types.ImageGenerateResponse{}, errors.New("images API returned empty data")
	}

	outDir := filepath.Join(p.staticRoot, "generated")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return types.ImageGenerateResponse{}, err
	}

	var urls []string
	for i, d := range out.Data {
		if d.URL != "" {
			urls = append(urls, d.URL)
			continue
		}
		if d.B64JSON == "" {
			continue
		}
		imgBytes, err := base64.StdEncoding.DecodeString(d.B64JSON)
		if err != nil {
			return types.ImageGenerateResponse{}, err
		}
		name := fmt.Sprintf("img_%d_%d.png", time.Now().UnixNano(), i)
		if err := os.WriteFile(filepath.Join(outDir, name), imgBytes, 0o644); err != nil {
			return types.ImageGenerateResponse{}, err
		}
		urls = append(urls, "/static/generated/"+name)
	}

	if len(urls) == 0 {
		return types.ImageGenerateResponse{}, errors.New("images API returned no usable images")
	}

	return types.ImageGenerateResponse{
		ImageURLs: urls,
		Provider:  p.ProviderName(),
	}, nil
}
