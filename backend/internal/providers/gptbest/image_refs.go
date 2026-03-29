package gptbest

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func mergeImageRefs(groups ...[]string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	for _, g := range groups {
		for _, raw := range g {
			ref := strings.TrimSpace(raw)
			if ref == "" {
				continue
			}
			if _, ok := seen[ref]; ok {
				continue
			}
			seen[ref] = struct{}{}
			out = append(out, ref)
		}
	}
	return out
}

func (p *Provider) loadRefImage(ctx context.Context, ref string, idx int) (string, []byte, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", nil, errors.New("empty image ref")
	}
	if strings.HasPrefix(strings.ToLower(ref), "data:image/") {
		ext, b, err := decodeImageDataURL(ref)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("image_%d.%s", idx+1, ext), b, nil
	}
	if !strings.HasPrefix(ref, "http://") && !strings.HasPrefix(ref, "https://") {
		return "", nil, errors.New("image ref must be data URL or HTTP URL")
	}

	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 25<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("download image failed: status=%d", resp.StatusCode)
	}
	ct := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	ext := extFromContentType(ct)
	return fmt.Sprintf("image_%d.%s", idx+1, ext), b, nil
}

func decodeImageDataURL(dataURL string) (string, []byte, error) {
	i := strings.Index(dataURL, ",")
	if i <= 0 {
		return "", nil, errors.New("invalid image data url")
	}
	head := strings.ToLower(strings.TrimSpace(dataURL[:i]))
	raw := strings.TrimSpace(dataURL[i+1:])
	if !strings.Contains(head, ";base64") {
		return "", nil, errors.New("image data url is not base64")
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", nil, err
	}
	return extFromContentType(head), b, nil
}

func extFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "image/jpeg"), strings.Contains(contentType, "image/jpg"):
		return "jpg"
	case strings.Contains(contentType, "image/webp"):
		return "webp"
	case strings.Contains(contentType, "image/gif"):
		return "gif"
	default:
		return "png"
	}
}
