package jeniya

import (
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/providers/gptbest"
	"aigc-backend/internal/providers/openai_compatible"
)

type Provider struct {
	baseURL string
	apiKey  string

	image *openai_compatible.Provider
	audio *gptbest.Provider
	http  *http.Client
}

func New(baseURL, apiKey, imageModel, staticRoot string) *Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://jeniya.cn"
	}
	apiKey = strings.TrimSpace(apiKey)
	return &Provider{
		baseURL: baseURL,
		apiKey:  apiKey,
		image:   openai_compatible.New(baseURL, apiKey, imageModel, staticRoot),
		audio:   gptbest.New("jeniya", baseURL, apiKey, "", staticRoot),
		http:    &http.Client{Timeout: 10 * time.Minute},
	}
}

func (p *Provider) ProviderName() string { return "jeniya" }
