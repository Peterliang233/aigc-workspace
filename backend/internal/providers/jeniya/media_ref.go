package jeniya

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (p *Provider) loadImage(ctx context.Context, ref string) (string, []byte, error) {
	ref = strings.TrimSpace(ref)
	switch {
	case ref == "":
		return "", nil, errors.New("empty image ref")
	case strings.HasPrefix(strings.ToLower(ref), "data:image/"):
		return decodeDataURL(ref)
	case strings.HasPrefix(ref, "http://"), strings.HasPrefix(ref, "https://"):
		return p.loadRemote(ctx, ref)
	default:
		return "", nil, errors.New("image ref must be data URL or HTTP URL")
	}
}

func (p *Provider) loadRemote(ctx context.Context, ref string) (string, []byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	resp, err := p.http.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("download image failed: status=%d", resp.StatusCode)
	}
	buf, err := io.ReadAll(io.LimitReader(resp.Body, 25<<20))
	if err != nil {
		return "", nil, err
	}
	return "reference." + extFromType(resp.Header.Get("Content-Type")), buf, nil
}

func decodeDataURL(ref string) (string, []byte, error) {
	idx := strings.Index(ref, ",")
	if idx <= 0 {
		return "", nil, errors.New("invalid image data url")
	}
	head := strings.ToLower(strings.TrimSpace(ref[:idx]))
	if !strings.Contains(head, ";base64") {
		return "", nil, errors.New("image data url is not base64")
	}
	buf, err := base64.StdEncoding.DecodeString(strings.TrimSpace(ref[idx+1:]))
	if err != nil {
		return "", nil, err
	}
	return "reference." + extFromType(head), buf, nil
}

func extFromType(s string) string {
	switch {
	case strings.Contains(strings.ToLower(s), "jpeg"), strings.Contains(strings.ToLower(s), "jpg"):
		return "jpg"
	case strings.Contains(strings.ToLower(s), "webp"):
		return "webp"
	case strings.Contains(strings.ToLower(s), "gif"):
		return "gif"
	default:
		return "png"
	}
}
