package siliconflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"aigc-backend/internal/util/mediafetch"
)

func (p *Provider) downloadAndStore(ctx context.Context, rawURL, outDir string, idx int) (string, error) {
	pu, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if pu.Scheme != "https" && pu.Scheme != "http" {
		return "", errors.New("refusing to download non-http(s) url")
	}

	ext := mediafetch.GuessExt("", rawURL)
	if ext == "" {
		ext = ".img"
	}

	name := fmt.Sprintf("sf_%d_%d_%s%s", time.Now().UnixNano(), idx, randHex(6), ext)
	dl := &mediafetch.Downloader{HTTP: p.httpClient}
	dst, _, err := dl.DownloadToDirAutoExt(ctx, rawURL, outDir, strings.TrimSuffix(name, ext), 25<<20)
	if err != nil {
		return "", err
	}
	return dst, nil
}

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

