package mediafetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
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
		d.HTTP = &http.Client{Timeout: 2 * time.Minute}
	}
	if maxBytes <= 0 {
		maxBytes = 25 << 20
	}

	pu, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}
	if pu.Scheme != "https" && pu.Scheme != "http" {
		return "", nil, errors.New("refusing to download non-http(s) url")
	}
	host := strings.ToLower(pu.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return "", nil, errors.New("refusing to download from localhost")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
			return "", nil, errors.New("refusing to download from local/private ip")
		}
	}

	dlCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
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

func GuessExt(contentType, rawURL string) string {
	if contentType != "" {
		mt, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			switch strings.ToLower(mt) {
			case "image/png":
				return ".png"
			case "image/jpeg":
				return ".jpg"
			case "image/webp":
				return ".webp"
			case "image/gif":
				return ".gif"
			case "image/svg+xml":
				return ".svg"
			}
		}
	}

	pu, err := url.Parse(rawURL)
	if err == nil {
		ext := strings.ToLower(filepath.Ext(pu.Path))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg":
			if ext == ".jpeg" {
				return ".jpg"
			}
			return ext
		}
	}
	return ""
}
