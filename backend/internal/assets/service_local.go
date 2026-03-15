package assets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/util/mediafetch"

	"github.com/minio/minio-go/v7"
)

func (s *Service) StoreLocalFile(ctx context.Context, in StoreLocalFileInput) (*Asset, error) {
	if !s.Enabled() {
		return nil, errors.New("asset storage is not configured")
	}
	if s.MaxBytes <= 0 {
		s.MaxBytes = 250 << 20 // 250MB default cap
	}

	in.Capability = strings.ToLower(strings.TrimSpace(in.Capability))
	in.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	in.Model = strings.TrimSpace(in.Model)
	in.FilePath = strings.TrimSpace(in.FilePath)
	in.ContentType = strings.TrimSpace(in.ContentType)

	if in.Capability == "" {
		return nil, errors.New("missing capability")
	}
	if in.Provider == "" {
		return nil, errors.New("missing provider")
	}
	if in.FilePath == "" {
		return nil, errors.New("missing file_path")
	}

	opCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	f, err := os.Open(in.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	if size > s.MaxBytes {
		return nil, fmt.Errorf("asset too large: %d bytes", size)
	}

	ct := in.ContentType
	if ct == "" {
		var buf [512]byte
		n, _ := io.ReadFull(f, buf[:])
		ct = http.DetectContentType(buf[:n])
		_, _ = f.Seek(0, 0)
	}
	if ct == "" {
		ct = "application/octet-stream"
	}

	ext := strings.TrimSpace(filepath.Ext(in.FilePath))
	if ext == "" {
		ext = mediafetch.GuessExt(ct, "")
	}
	if ext == "" {
		if exts, _ := mime.ExtensionsByType(ct); len(exts) > 0 {
			ext = exts[0]
		}
	}
	if ext == "" {
		ext = ".bin"
	}

	objectKey, err := makeObjectKey(in.Capability, ext)
	if err != nil {
		return nil, err
	}

	if _, err := s.MinIO.Client.PutObject(opCtx, s.MinIO.Bucket, objectKey, io.LimitReader(f, s.MaxBytes+1), size, minio.PutObjectOptions{
		ContentType: ct,
	}); err != nil {
		return nil, err
	}

	sha, preview := promptMeta(in.Prompt)
	a := &Asset{
		Capability:    in.Capability,
		Provider:      in.Provider,
		Model:         in.Model,
		PromptSHA256:  sha,
		PromptPreview: preview,
		ParamsJSON:    marshalParams(in.Params),
		Status:        "succeeded",
		ObjectKey:     objectKey,
		ContentType:   ct,
		Bytes:         size,
	}

	if _, err := s.Store.Create(opCtx, a); err != nil {
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
		"source", "local_file",
	)
	return a, nil
}

