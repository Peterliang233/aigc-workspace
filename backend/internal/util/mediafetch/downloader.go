package mediafetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/logging"
)

type Downloader struct {
	HTTP *http.Client
}

func (d *Downloader) DownloadToFile(ctx context.Context, rawURL, dstPath string, maxBytes int64) error {
	_, _, err := d.download(ctx, rawURL, dstPath, maxBytes, false)
	return err
}

// DownloadToDirAutoExt downloads the URL and stores it under dir using a detected file extension.
// It returns the absolute dst path and the detected content-type.
func (d *Downloader) DownloadToDirAutoExt(ctx context.Context, rawURL, dir, namePrefix string, maxBytes int64) (string, string, error) {
	if strings.TrimSpace(namePrefix) == "" {
		namePrefix = "file"
	}
	tmp := filepath.Join(dir, namePrefix+".tmp")
	ct, data, err := d.download(ctx, rawURL, tmp, maxBytes, true)
	if err != nil {
		_ = os.Remove(tmp)
		return "", "", err
	}

	ext := GuessExt(ct, rawURL)
	if ext == "" {
		ext = GuessExt(http.DetectContentType(data), rawURL)
	}
	if ext == "" {
		ext = ".bin"
	}

	dst := filepath.Join(dir, namePrefix+ext)
	if err := os.Rename(tmp, dst); err != nil {
		// Cross-device rename fallback.
		if err2 := os.WriteFile(dst, data, 0o644); err2 != nil {
			_ = os.Remove(tmp)
			return "", "", err2
		}
		_ = os.Remove(tmp)
	}

	return dst, ct, nil
}

func (d *Downloader) download(ctx context.Context, rawURL, dstPath string, maxBytes int64, returnBytes bool) (string, []byte, error) {
	if d.HTTP == nil {
		d.HTTP = &http.Client{Timeout: 10 * time.Minute}
	}
	if maxBytes <= 0 {
		maxBytes = 25 << 20
	}

	if err := validateURL(rawURL); err != nil {
		return "", nil, err
	}

	dlCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	slog.Default().Debug("download_start",
		"url", logging.RedactURL(rawURL),
		"dst", dstPath,
		"max_bytes", maxBytes,
	)
	hreq, _ := http.NewRequestWithContext(dlCtx, http.MethodGet, rawURL, nil)
	resp, err := d.HTTP.Do(hreq)
	if err != nil {
		slog.Default().Warn("download_failed", "url", logging.RedactURL(rawURL), "err", err.Error())
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		slog.Default().Warn("download_bad_status",
			"url", logging.RedactURL(rawURL),
			"status", resp.StatusCode,
		)
		return "", nil, fmt.Errorf("download error: status=%d body=%s", resp.StatusCode, string(b))
	}

	r := io.LimitReader(resp.Body, maxBytes+1)
	b, err := io.ReadAll(r)
	if err != nil {
		return "", nil, err
	}
	if int64(len(b)) > maxBytes {
		return "", nil, errors.New("downloaded file too large")
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(dstPath, b, 0o644); err != nil {
		slog.Default().Warn("download_write_failed", "dst", dstPath, "err", err.Error())
		return "", nil, err
	}

	slog.Default().Debug("download_ok",
		"url", logging.RedactURL(rawURL),
		"dst", dstPath,
		"bytes", len(b),
		"ct", resp.Header.Get("Content-Type"),
	)

	if returnBytes {
		return resp.Header.Get("Content-Type"), b, nil
	}
	return resp.Header.Get("Content-Type"), nil, nil
}

// Open issues a GET request for a remote URL after applying basic SSRF protections.
// Caller must close resp.Body.
func (d *Downloader) Open(ctx context.Context, rawURL string) (*http.Response, error) {
	if d.HTTP == nil {
		d.HTTP = &http.Client{Timeout: 10 * time.Minute}
	}
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}

	slog.Default().Debug("download_open",
		"url", logging.RedactURL(rawURL),
	)

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	resp, err := d.HTTP.Do(hreq)
	if err != nil {
		slog.Default().Warn("download_open_failed", "url", logging.RedactURL(rawURL), "err", err.Error())
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download error: status=%d body=%s", resp.StatusCode, string(b))
	}
	return resp, nil
}

