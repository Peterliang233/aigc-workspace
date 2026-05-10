package textgen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
)

type Request struct {
	Model        string
	SystemPrompt string
	Prompt       string
	Temperature  *float64
	MaxTokens    *int
}

type Response struct {
	Text     string
	Provider string
	Model    string
}

type Client struct {
	provider   string
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func New(provider, baseURL, apiKey, model string) *Client {
	return &Client{
		provider:   strings.ToLower(strings.TrimSpace(provider)),
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *Client) Enabled() bool { return c != nil && c.baseURL != "" && c.apiKey != "" }

func (c *Client) Generate(ctx context.Context, req Request) (Response, error) {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = c.model
	}
	body, _ := buildRequest(model, req)
	url := c.baseURL + "/v1/chat/completions"
	logging.DownstreamRequestRaw("text_generate_request", c.provider, http.MethodPost, url, map[string]any{
		"model":         model,
		"system_prompt": logging.DownstreamPrompt(req.SystemPrompt),
		"prompt":        logging.DownstreamPrompt(req.Prompt),
	}, "application/json", body)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	hreq.Header.Set("Authorization", "Bearer "+c.apiKey)
	hreq.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := c.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseRaw("text_generate_response", c.provider, http.MethodPost, url, 0, time.Since(start), err, "", nil)
		return Response{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseRaw("text_generate_response", c.provider, http.MethodPost, url, resp.StatusCode, time.Since(start), fmt.Errorf("bad status"), resp.Header.Get("Content-Type"), raw)
		return Response{}, fmt.Errorf("text API error: status=%d body=%s", resp.StatusCode, string(raw))
	}
	logging.DownstreamResponseRaw("text_generate_response", c.provider, http.MethodPost, url, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), raw)
	text, err := parseResponse(raw)
	if err != nil {
		return Response{}, err
	}
	return Response{Text: text, Provider: c.provider, Model: model}, nil
}

func buildRequest(model string, req Request) ([]byte, error) {
	messages := []map[string]string{}
	if strings.TrimSpace(req.SystemPrompt) != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.SystemPrompt})
	}
	messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})
	payload := map[string]any{"model": model, "messages": messages}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		payload["max_tokens"] = *req.MaxTokens
	}
	return json.Marshal(payload)
}
