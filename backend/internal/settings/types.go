package settings

// ProviderSettings is persisted configuration. Pointers allow tri-state:
// - nil: not set (inherit from env/defaults)
// - non-nil: set explicitly (including empty string to clear)
type ProviderSettings struct {
	BaseURL      *string   `json:"base_url,omitempty"`
	APIKey       *string   `json:"api_key,omitempty"`
	DefaultModel *string   `json:"default_model,omitempty"`
	Models       *[]string `json:"models,omitempty"`
}

type Settings struct {
	ImageProviders map[string]ProviderSettings `json:"image_providers,omitempty"`
}
