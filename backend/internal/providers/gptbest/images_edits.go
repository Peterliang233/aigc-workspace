package gptbest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

type imageAPIResp struct {
	Data []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

func (p *Provider) generateByEdits(
	ctx context.Context,
	model, prompt string,
	req types.ImageGenerateRequest,
	refs []string,
) (types.ImageGenerateResponse, error) {
	payload, ctype, err := p.buildEditsPayload(ctx, model, prompt, req, refs)
	if err != nil {
		return types.ImageGenerateResponse{}, err
	}
	u := imagesEditsURL(p.baseURL)
	logging.DownstreamRequestRaw("provider_gptbest_request", p.ProviderName(), http.MethodPost, u, map[string]any{
		"model":           model,
		"image_count":     len(refs),
		"size":            strings.TrimSpace(req.Size),
		"aspect_ratio":    strings.TrimSpace(req.AspectRatio),
		"negative_prompt": logging.DownstreamPrompt(req.NegativePrompt),
		"prompt":          logging.DownstreamPrompt(prompt),
	}, ctype, payload)

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	hreq.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", ctype)

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

func (p *Provider) buildEditsPayload(
	ctx context.Context,
	model, prompt string,
	req types.ImageGenerateRequest,
	refs []string,
) ([]byte, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("model", model)
	_ = w.WriteField("prompt", prompt)
	if s := strings.TrimSpace(req.Size); s != "" {
		_ = w.WriteField("size", s)
	}
	if s := strings.TrimSpace(req.AspectRatio); s != "" {
		_ = w.WriteField("aspect_ratio", s)
	}
	if s := strings.TrimSpace(req.NegativePrompt); s != "" {
		_ = w.WriteField("negative_prompt", s)
	}
	if s := strings.TrimSpace(req.Style); s != "" {
		_ = w.WriteField("style", s)
	}
	if req.Seed != nil {
		_ = w.WriteField("seed", strconv.FormatInt(*req.Seed, 10))
	}
	if req.Strength != nil {
		_ = w.WriteField("strength", strconv.FormatFloat(*req.Strength, 'f', -1, 64))
	}
	for i, ref := range refs {
		name, b, err := p.loadRefImage(ctx, ref, i)
		if err != nil {
			return nil, "", err
		}
		part, err := w.CreateFormFile("image", name)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(b); err != nil {
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

func (p *Provider) parseImageAPIResponse(body []byte) ([]string, error) {
	var out imageAPIResp
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(out.Data))
	for _, d := range out.Data {
		if s := strings.TrimSpace(d.URL); s != "" {
			urls = append(urls, s)
			continue
		}
		if s := strings.TrimSpace(d.B64JSON); s != "" {
			u, err := p.storeBase64Image(s, "png")
			if err != nil {
				return nil, err
			}
			urls = append(urls, u)
		}
	}
	if len(urls) == 0 {
		return nil, errors.New("provider returned no image url")
	}
	return urls, nil
}

func imagesEditsURL(base string) string {
	u := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(strings.ToLower(u), "/v1/images/edits") {
		return u
	}
	return u + "/v1/images/edits"
}
