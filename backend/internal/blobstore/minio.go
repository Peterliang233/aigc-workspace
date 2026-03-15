package blobstore

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIO struct {
	Client *minio.Client
	Bucket string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func NewMinIO(cfg MinIOConfig) (*MinIO, error) {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)

	if cfg.Endpoint == "" {
		return nil, errors.New("MINIO_ENDPOINT is empty")
	}
	if cfg.AccessKey == "" {
		return nil, errors.New("MINIO_ACCESS_KEY is empty")
	}
	if cfg.SecretKey == "" {
		return nil, errors.New("MINIO_SECRET_KEY is empty")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("MINIO_BUCKET is empty")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}

	cl, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.UseSSL,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}

	m := &MinIO{Client: cl, Bucket: cfg.Bucket}
	if err := m.ensureBucketWithRetry(60 * time.Second); err != nil {
		return nil, err
	}

	slog.Default().Info("minio_ready",
		"endpoint", cfg.Endpoint,
		"bucket", cfg.Bucket,
		"ssl", cfg.UseSSL,
	)
	return m, nil
}

func (m *MinIO) ensureBucketWithRetry(maxWait time.Duration) error {
	start := time.Now()
	sleep := 250 * time.Millisecond

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		ok, err := m.Client.BucketExists(ctx, m.Bucket)
		cancel()

		if err == nil {
			if ok {
				return nil
			}
			ctx2, cancel2 := context.WithTimeout(context.Background(), 8*time.Second)
			err = m.Client.MakeBucket(ctx2, m.Bucket, minio.MakeBucketOptions{})
			cancel2()
			if err == nil {
				slog.Default().Info("minio_bucket_created", "bucket", m.Bucket)
				return nil
			}
			// If another initializer created it in parallel, treat as OK.
			if strings.Contains(strings.ToLower(err.Error()), "bucketalready") {
				return nil
			}
		}

		if time.Since(start) >= maxWait {
			if err != nil {
				return fmt.Errorf("minio not ready (bucket=%s): %w", m.Bucket, err)
			}
			return fmt.Errorf("minio not ready (bucket=%s)", m.Bucket)
		}

		slog.Default().Warn("minio_wait",
			"bucket", m.Bucket,
			"elapsed_ms", time.Since(start).Milliseconds(),
		)
		time.Sleep(sleep)
		if sleep < 2*time.Second {
			sleep *= 2
		}
	}
}

func (m *MinIO) Redacted() map[string]any {
	if m == nil || m.Client == nil {
		return map[string]any{"enabled": false}
	}
	// Endpoint isn't a secret, but avoid printing full URLs with credentials (if any) by keeping it host-only.
	endpoint := ""
	if u := m.Client.EndpointURL(); u != nil {
		endpoint = u.Host
	}
	return map[string]any{"enabled": true, "bucket": m.Bucket, "endpoint": endpoint}
}
