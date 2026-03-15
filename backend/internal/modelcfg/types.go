package modelcfg

// Config is a repo-managed JSON file that defines:
// - which providers/models to show in the UI
// - per-model form requirements (e.g. I2V requires an init image)
//
// Secrets (API keys) and base URLs are NOT stored here.
type Config struct {
	Version   int        `json:"version"`
	Defaults  Defaults   `json:"defaults"`
	Providers []Provider `json:"providers"`
}

type Defaults struct {
	ImageProvider string `json:"image_provider"`
	VideoProvider string `json:"video_provider"`
}

type Provider struct {
	ID    string          `json:"id"`
	Label string          `json:"label"`
	Image *CapabilitySpec `json:"image,omitempty"`
	Video *CapabilitySpec `json:"video,omitempty"`
}

type CapabilitySpec struct {
	DefaultModel string      `json:"default_model"`
	Models       []ModelSpec `json:"models"`
}

type ModelSpec struct {
	ID    string    `json:"id"`
	Label string    `json:"label,omitempty"`
	Form  *FormSpec `json:"form,omitempty"`
}

type FormSpec struct {
	RequiresImage bool `json:"requires_image,omitempty"`
}
