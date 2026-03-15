package wuyinkeji

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

// Implements the "速创API-图片模型-图片生成" family.
// Example doc: https://api.wuyinkeji.com/doc/65 (NanoBanana2)
//
// Their image generation endpoints are async:
// - POST /api/async/<model_endpoint>?key=API_KEY  => {data:{id:"..."}}
// - GET  /api/async/detail?key=API_KEY&id=...     => {data:{status:2, ... result ...}}
//
// We wrap it as a "sync" call to match this app's /api/images/generate:
// start job -> poll -> extract image URLs -> download to /static/ (best effort).
type Provider struct {
	baseURL string
	apiKey  string

	// Optional allow-list of models shown in UI. If non-empty, requests must pick one of them.
	// For "dynamic endpoint" mode, each model is the path segment in:
	//   POST /api/async/{model}?key=...
	models []string

	// Legacy mode: model label -> endpoint path (starting with /).
	// If configured, we use this mapping instead of the dynamic endpoint pattern.
	endpoints map[string]string

	httpClient *http.Client
	staticRoot string
}

func New(baseURL, apiKey, staticRoot string, models []string, endpoints map[string]string) *Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.wuyinkeji.com"
	}

	var ms []string
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		// normalize common full paths to "segment" form
		m = strings.TrimPrefix(m, "/api/async/")
		m = strings.TrimPrefix(m, "api/async/")
		ms = append(ms, m)
	}

	ep := map[string]string{}
	for k, v := range endpoints {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		if !strings.HasPrefix(v, "/") {
			v = "/" + v
		}
		ep[k] = v
	}

	return &Provider{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		models:     ms,
		endpoints:  ep,
		httpClient: &http.Client{Timeout: 10 * time.Minute},
		staticRoot: staticRoot,
	}
}

func (p *Provider) ProviderName() string { return "wuyinkeji" }

func (p *Provider) Models() []string {
	// Prefer legacy mapping keys if present; otherwise return configured model segments.
	if len(p.endpoints) > 0 {
		var out []string
		for m := range p.endpoints {
			out = append(out, m)
		}
		sort.Strings(out)
		return out
	}
	out := make([]string, 0, len(p.models))
	out = append(out, p.models...)
	sort.Strings(out)
	return out
}

