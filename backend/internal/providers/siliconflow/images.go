package siliconflow

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
	"aigc-backend/internal/util/mediafetch"
)

// Implements https://docs.siliconflow.cn/cn/api-reference/images/images-generations
// SiliconFlow returns temporary image URLs (typically valid for ~1 hour), so we download
// and store them under /static/ for reliable rendering in the web app.
type Provider struct {
	baseURL      string
	apiKey       string
	defaultModel string

	httpClient *http.Client

	staticRoot string
}

func New(baseURL, apiKey, defaultModel, staticRoot string) *Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.siliconflow.cn"
	}
	return &Provider{
		baseURL:      baseURL,
		apiKey:       strings.TrimSpace(apiKey),
		defaultModel: strings.TrimSpace(defaultModel),
		httpClient:   &http.Client{Timeout: 10 * time.Minute},
		staticRoot:   staticRoot,
	}
}

func (p *Provider) ProviderName() string { return "siliconflow" }

type imagesGenerateRequest struct {
	Model             string  `json:"model"`
	Prompt            string  `json:"prompt"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	ImageSize         string  `json:"image_size"`
	BatchSize         int     `json:"batch_size,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
}

type imagesGenerateResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

func (p *Provider) GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error) {
	if p.apiKey == "" {
		return types.ImageGenerateResponse{}, errors.New("SiliconFlow 未配置 API Key")
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return types.ImageGenerateResponse{}, errors.New("prompt is required")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.defaultModel
	}
	if model == "" {
		return types.ImageGenerateResponse{}, errors.New("SiliconFlow 需要指定模型")
	}

	imageSize := strings.TrimSpace(req.Size)
	if imageSize == "" {
		imageSize = "1024x1024"
	}

	batch := req.N
	if batch <= 0 {
		batch = 1
	}

	body := imagesGenerateRequest{
		Model:             model,
		Prompt:            prompt,
		ImageSize:         imageSize,
		BatchSize:         batch,
		NumInferenceSteps: 20,
		GuidanceScale:     7.5,
	}

	raw, _ := json.Marshal(body)
	u := p.baseURL + "/v1/images/generations"
	logging.DownstreamRequest("provider_siliconflow_request", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":               model,
		"image_size":          imageSize,
		"batch_size":          batch,
		"num_inference_steps": body.NumInferenceSteps,
		"guidance_scale":      body.GuidanceScale,
		"prompt":              logging.DownstreamPrompt(prompt),
	})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponse("provider_siliconflow_response", p.ProviderName(), http.MethodPost, u, 0, time.Since(start), err)
		return types.ImageGenerateResponse{}, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponse("provider_siliconflow_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), errors.New("bad status"))
		slog.Default().Warn("provider_siliconflow_error", "status", resp.StatusCode)
		return types.ImageGenerateResponse{}, fmt.Errorf("siliconflow images API error: status=%d body=%s", resp.StatusCode, string(b))
	}
	logging.DownstreamResponse("provider_siliconflow_response", p.ProviderName(), http.MethodPost, u, resp.StatusCode, time.Since(start), nil)

	var out imagesGenerateResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return types.ImageGenerateResponse{}, err
	}
	if len(out.Images) == 0 {
		slog.Default().Warn("provider_siliconflow_no_images")
		return types.ImageGenerateResponse{}, errors.New("siliconflow returned empty images")
	}

	outDir := filepath.Join(p.staticRoot, "generated")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return types.ImageGenerateResponse{}, err
	}

	var urls []string
	for i, img := range out.Images {
		remote := strings.TrimSpace(img.URL)
		if remote == "" {
			continue
		}
		local, err := p.downloadAndStore(ctx, remote, outDir, i)
		if err != nil {
			slog.Default().Warn("provider_siliconflow_download_failed", "err", err.Error())
			// Fallback: at least return the remote URL (may expire).
			urls = append(urls, remote)
			continue
		}
		urls = append(urls, "/static/generated/"+filepath.Base(local))
	}

	if len(urls) == 0 {
		return types.ImageGenerateResponse{}, errors.New("siliconflow returned no usable image urls")
	}

	return types.ImageGenerateResponse{
		ImageURLs: urls,
		Provider:  p.ProviderName(),
		Model:     model,
	}, nil
}

func (p *Provider) downloadAndStore(ctx context.Context, rawURL, outDir string, idx int) (string, error) {
	pu, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if pu.Scheme != "https" && pu.Scheme != "http" {
		return "", errors.New("refusing to download non-http(s) url")
	}

	ext := guessImageExt("", rawURL)
	if ext == "" {
		ext = ".img"
	}

	name := fmt.Sprintf("sf_%d_%d_%s%s", time.Now().UnixNano(), idx, randHex(6), ext)
	dl := &mediafetch.Downloader{HTTP: p.httpClient}
	dst, _, err := dl.DownloadToDirAutoExt(ctx, rawURL, outDir, strings.TrimSuffix(name, ext), 25<<20)
	if err != nil {
		return "", err
	}
	return dst, nil
}

func guessImageExt(contentType, rawURL string) string {
	return mediafetch.GuessExt(contentType, rawURL)
}

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
