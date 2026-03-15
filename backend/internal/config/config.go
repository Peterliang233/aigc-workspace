package config

import (
	"os"
	"strings"
)

type Config struct {
	Provider       string
	BaseURL        string
	APIKey         string
	ImageModel     string
	VideoModel     string
	VideoStartEP   string
	VideoStatusEP  string
	Port           string
	AllowedOrigins []string
}

func LoadFromEnv() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	origins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	var allowed []string
	if origins != "" {
		for _, o := range strings.Split(origins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowed = append(allowed, o)
			}
		}
	}
	if len(allowed) == 0 {
		allowed = []string{"http://localhost:5173"}
	}

	provider := strings.TrimSpace(os.Getenv("AIGC_PROVIDER"))
	if provider == "" {
		provider = "mock"
	}

	return Config{
		Provider:       provider,
		BaseURL:        strings.TrimSpace(os.Getenv("AIGC_BASE_URL")),
		APIKey:         strings.TrimSpace(os.Getenv("AIGC_API_KEY")),
		ImageModel:     strings.TrimSpace(os.Getenv("AIGC_IMAGE_MODEL")),
		VideoModel:     strings.TrimSpace(os.Getenv("AIGC_VIDEO_MODEL")),
		VideoStartEP:   strings.TrimSpace(os.Getenv("AIGC_VIDEO_START_ENDPOINT")),
		VideoStatusEP:  strings.TrimSpace(os.Getenv("AIGC_VIDEO_STATUS_ENDPOINT")),
		Port:           port,
		AllowedOrigins: allowed,
	}
}
