package types

type ImageGenerateRequest struct {
	Prompt string `json:"prompt"`
	Size   string `json:"size,omitempty"` // e.g. 1024x1024
	N      int    `json:"n,omitempty"`
}

type ImageGenerateResponse struct {
	ImageURLs []string `json:"image_urls"`
	Provider  string   `json:"provider"`
}

type VideoJobCreateRequest struct {
	Prompt          string `json:"prompt"`
	DurationSeconds int    `json:"duration_seconds,omitempty"`
	AspectRatio     string `json:"aspect_ratio,omitempty"` // e.g. 16:9
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
