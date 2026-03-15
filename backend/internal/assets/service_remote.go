package assets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"aigc-backend/internal/logging"
	"aigc-backend/internal/util/mediafetch"

	"github.com/minio/minio-go/v7"
)

func (s *Service) StoreRemote(ctx context.Context, in StoreRemoteInput) (*Asset, error) {
	if !s.Enabled() {
		return nil, errors.New("asset storage is not configured")
	}
	if s.Download == nil {
		s.Download = &mediafetch.Downloader{}
	}
	if s.MaxBytes <= 0 {
		s.MaxBytes = 250 << 20 // 250MB default cap
	}

	in.Capability = strings.ToLower(strings.TrimSpace(in.Capability))
	in.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	in.Model = strings.TrimSpace(in.Model)
	in.SourceURL = strings.TrimSpace(in.SourceURL)
	in.ExternalJobID = strings.TrimSpace(in.ExternalJobID)
	if in.Capability == "" {
		return nil, errors.New("missing capability")
	}
	if in.Provider == "" {
		return nil, errors.New("missing provider")
	}
	if in.SourceURL == "" {
		return nil, errors.New("missing source_url")
	}

	if in.ExternalJobID != "" {
		if existing, err := s.Store.FindByExternalJobID(ctx, in.ExternalJobID); err != nil {
			return nil, err
		} else if existing != nil {
			return existing, nil
		}
	}

	// 10 minutes total budget for fetching + storing the result object.
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	resp, err := s.Download.Open(opCtx, in.SourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ct := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if ct == "" {
		ct = "application/octet-stream"
	}
	ext := mediafetch.GuessExt(ct, in.SourceURL)
	if ext == "" {
		ext = ".bin"
	}

	objectKey, err := makeObjectKey(in.Capability, ext)
	if err != nil {
		return nil, err
	}

	size := resp.ContentLength
	if size > s.MaxBytes {
		return nil, fmt.Errorf("asset too large: %d bytes", size)
	}

	put := func(r io.Reader, size int64) error {
		_, err := s.MinIO.Client.PutObject(opCtx, s.MinIO.Bucket, objectKey, r, size, minio.PutObjectOptions{
			ContentType: ct,
		})
		return err
	}

	// If server didn't include Content-Length, spool to a temp file so we can upload with a known size.
	if size < 0 {
		tmpDir := os.TempDir()
		f, err := os.CreateTemp(tmpDir, "aigc-asset-*"+ext)
		if err != nil {
			return nil, err
		}
		tmpPath := f.Name()
		defer func() {
			_ = f.Close()
			_ = os.Remove(tmpPath)
		}()

		// Enforce max size while spooling.
		n, err := io.Copy(f, io.LimitReader(resp.Body, s.MaxBytes+1))
		if err != nil {
			return nil, err
		}
		if n > s.MaxBytes {
			return nil, fmt.Errorf("asset too large: %d bytes", n)
		}
		if _, err := f.Seek(0, 0); err != nil {
			return nil, err
		}
		size = n
		if err := put(f, size); err != nil {
			return nil, err
		}
	} else {
		// Known size: stream directly, but still enforce upper bound.
		if err := put(io.LimitReader(resp.Body, s.MaxBytes+1), size); err != nil {
			return nil, err
		}
	}

	sha, preview := promptMeta(in.Prompt)
	redactedURL := logging.RedactURL(in.SourceURL)
	a := &Asset{
		Capability:    in.Capability,
		Provider:      in.Provider,
		Model:         in.Model,
		PromptSHA256:  sha,
		PromptPreview: preview,
		ParamsJSON:    marshalParams(in.Params),
		Status:        "succeeded",
		SourceURL:     &redactedURL,
		ObjectKey:     objectKey,
		ContentType:   ct,
		Bytes:         size,
	}
	if in.ExternalJobID != "" {
		a.ExternalJobID = &in.ExternalJobID
	}

	if _, err := s.Store.Create(opCtx, a); err != nil {
		// Upload already succeeded; keep object for debugging and surface DB error.
		slog.Default().Error("asset_db_create_failed",
			"capability", in.Capability,
			"provider", in.Provider,
			"object_key", objectKey,
			"bytes", size,
			"err", err.Error(),
		)
		return nil, err
	}

	slog.Default().Info("asset_stored",
		"id", a.ID,
		"capability", a.Capability,
		"provider", a.Provider,
		"model", a.Model,
		"object_key", a.ObjectKey,
		"bytes", a.Bytes,
		"ct", a.ContentType,
		"source_url", logging.RedactURL(in.SourceURL),
	)
	return a, nil
}

