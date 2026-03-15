package config

import (
	"context"
	"net"
	"os"
	"strings"
	"time"
)

type ImageProviderConfig struct {
	BaseURL       string
	APIKey        string
	DefaultModel  string
	Models        []string
	ModelEndpoint map[string]string // for async APIs where model => endpoint path
}

type Config struct {
	Provider       string
	VideoModel     string
	VideoStartEP   string
	VideoStatusEP  string
	Port           string
	AllowedOrigins []string

	ImageProviders map[string]ImageProviderConfig

	// Settings are stored in MySQL (required).
	MySQLDSN string

	// Asset storage (MinIO/S3 compatible). Optional but recommended for history.
	MinIOEndpoint      string
	MinIOAccessKey     string
	MinIOSecretKey     string
	MinIOBucket        string
	MinIOUseSSL        bool
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

	// Backward-compatible shared envs (single-provider mode).
	sharedBase := strings.TrimSpace(os.Getenv("AIGC_BASE_URL"))
	sharedKey := strings.TrimSpace(os.Getenv("AIGC_API_KEY"))
	sharedModel := strings.TrimSpace(os.Getenv("AIGC_IMAGE_MODEL"))

	openaiBase := firstNonEmpty(strings.TrimSpace(os.Getenv("OPENAI_COMPAT_BASE_URL")), sharedBase)
	openaiKey := firstNonEmpty(strings.TrimSpace(os.Getenv("OPENAI_COMPAT_API_KEY")), sharedKey)
	openaiModel := firstNonEmpty(strings.TrimSpace(os.Getenv("OPENAI_COMPAT_IMAGE_MODEL")), sharedModel)
	openaiModels := parseCSV(os.Getenv("OPENAI_COMPAT_IMAGE_MODELS"))

	sfBase := firstNonEmpty(strings.TrimSpace(os.Getenv("SILICONFLOW_BASE_URL")), sharedBase)
	sfKey := firstNonEmpty(strings.TrimSpace(os.Getenv("SILICONFLOW_API_KEY")), sharedKey)
	sfModel := firstNonEmpty(strings.TrimSpace(os.Getenv("SILICONFLOW_IMAGE_MODEL")), sharedModel)
	sfModels := parseCSV(os.Getenv("SILICONFLOW_IMAGE_MODELS"))

	wyBase := strings.TrimSpace(os.Getenv("WUYIN_BASE_URL"))
	wyKey := strings.TrimSpace(os.Getenv("WUYIN_API_KEY"))
	wyModels := parseCSV(os.Getenv("WUYIN_IMAGE_MODELS"))
	wyEndpoints := parseKVCSV(os.Getenv("WUYIN_IMAGE_ENDPOINTS"))

	imageProviders := map[string]ImageProviderConfig{
		"mock": {
			BaseURL:      "",
			APIKey:       "",
			DefaultModel: "",
			Models:       nil,
		},
		"openai_compatible": {
			BaseURL:      openaiBase,
			APIKey:       openaiKey,
			DefaultModel: openaiModel,
			Models:       openaiModels,
		},
		"siliconflow": {
			BaseURL:      sfBase,
			APIKey:       sfKey,
			DefaultModel: sfModel,
			Models:       sfModels,
		},
		"wuyinkeji": {
			BaseURL:       wyBase,
			APIKey:        wyKey,
			DefaultModel:  "",
			Models:        wyModels,
			ModelEndpoint: wyEndpoints,
		},
	}

	return Config{
		Provider:       provider,
		VideoModel:     strings.TrimSpace(os.Getenv("AIGC_VIDEO_MODEL")),
		VideoStartEP:   strings.TrimSpace(os.Getenv("AIGC_VIDEO_START_ENDPOINT")),
		VideoStatusEP:  strings.TrimSpace(os.Getenv("AIGC_VIDEO_STATUS_ENDPOINT")),
		Port:           port,
		AllowedOrigins: allowed,
		ImageProviders: imageProviders,
		MySQLDSN:       pickMySQLDSN(),
		MinIOEndpoint:  pickMinIOEndpoint(),
		MinIOAccessKey: firstNonEmpty(strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY")), strings.TrimSpace(os.Getenv("MINIO_ROOT_USER"))),
		MinIOSecretKey: firstNonEmpty(strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY")), strings.TrimSpace(os.Getenv("MINIO_ROOT_PASSWORD"))),
		MinIOBucket:    strings.TrimSpace(os.Getenv("MINIO_BUCKET")),
		MinIOUseSSL:    strings.TrimSpace(os.Getenv("MINIO_USE_SSL")) == "true",
	}
}

func pickMySQLDSN() string {
	// In docker compose we typically use tcp(mysql:3306).
	// When running `go run` on the host, that hostname won't resolve, so we
	// allow a separate MYSQL_DSN_LOCAL (e.g. tcp(127.0.0.1:3307)).
	dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	local := strings.TrimSpace(os.Getenv("MYSQL_DSN_LOCAL"))
	if local == "" {
		return dsn
	}
	if dsn == "" {
		return local
	}
	if !strings.Contains(dsn, "@tcp(mysql:") {
		return dsn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, err := net.DefaultResolver.LookupHost(ctx, "mysql"); err != nil {
		return local
	}
	return dsn
}

func pickMinIOEndpoint() string {
	// In docker compose we typically use minio:9000.
	// When running `go run` on the host, that hostname won't resolve, so we
	// allow a separate MINIO_ENDPOINT_LOCAL (e.g. 127.0.0.1:9000).
	ep := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	local := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT_LOCAL"))
	if local == "" {
		return ep
	}
	if ep == "" {
		return local
	}
	if !strings.HasPrefix(strings.ToLower(ep), "minio:") && !strings.Contains(ep, "minio:") {
		return ep
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, err := net.DefaultResolver.LookupHost(ctx, "minio"); err != nil {
		return local
	}
	return ep
}

func parseCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseKVCSV(s string) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	return out
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
