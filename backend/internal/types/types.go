package types

type ImageGenerateRequest struct {
	// Provider/model are optional. If absent, backend uses its default provider config.
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`

	Prompt string `json:"prompt"`
	Size   string `json:"size,omitempty"` // e.g. 1024x1024
	N      int    `json:"n,omitempty"`

	// Optional, provider-specific extensions (safe to ignore by providers that don't support them).
	NegativePrompt string   `json:"negative_prompt,omitempty"`
	AspectRatio    string   `json:"aspect_ratio,omitempty"`   // e.g. 16:9
	Image          []string `json:"image,omitempty"`          // OpenAI-like i2i refs
	ReferenceURLs  []string `json:"reference_urls,omitempty"` // e.g. input images
	Seed           *int64   `json:"seed,omitempty"`           // optional seed
	Strength       *float64 `json:"strength,omitempty"`       // e.g. img2img strength
	Style          string   `json:"style,omitempty"`          // optional style
	Extra          any      `json:"extra,omitempty"`          // last resort for provider-specific payloads (must be JSON object)
}

type ImageGenerateResponse struct {
	ImageURLs []string `json:"image_urls"`
	Provider  string   `json:"provider"`
	Model     string   `json:"model,omitempty"`
}

type VideoJobCreateRequest struct {
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`

	Prompt          string `json:"prompt"`
	DurationSeconds int    `json:"duration_seconds,omitempty"`
	AspectRatio     string `json:"aspect_ratio,omitempty"` // legacy (generic providers)

	// SiliconFlow specific fields (safe to ignore by other providers).
	ImageSize      string         `json:"image_size,omitempty"`      // 1280x720|720x1280|960x960
	NegativePrompt string         `json:"negative_prompt,omitempty"` // optional
	Image          string         `json:"image,omitempty"`           // URL or base64 (I2V)
	Seed           *int64         `json:"seed,omitempty"`            // optional seed
	Extra          map[string]any `json:"extra,omitempty"`           // provider/model-specific passthrough fields
}

type VideoJobCreateResponse struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
}

type VideoJobGetResponse struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	VideoURL string `json:"video_url,omitempty"`
	Error    string `json:"error,omitempty"`
	Provider string `json:"provider"`
}
