package aliyunbailian

import (
	"net/http"
	"strings"
	"time"
)

type Provider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string) *Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/api/v1"
	}
	return &Provider{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (p *Provider) ProviderName() string { return "aliyunbailian" }

func (p *Provider) generationURL() string {
	const path = "/services/aigc/multimodal-generation/generation"
	base := strings.TrimRight(p.baseURL, "/")
	if strings.HasSuffix(base, path) {
		return base
	}
	if strings.HasSuffix(base, "/api/v1") {
		return base + path
	}
	return base + "/api/v1" + path
}
