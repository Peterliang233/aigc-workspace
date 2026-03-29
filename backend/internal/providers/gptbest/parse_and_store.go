package gptbest

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	reMarkdownImage = regexp.MustCompile(`!\[[^\]]*]\((https?://[^)\s]+)\)`)
	rePlainURL      = regexp.MustCompile(`https?://[^\s\])"'>]+`)
)

func extractImageRef(out chatResp) string {
	for _, c := range out.Choices {
		if u := extractFromAny(c.Message.Content); u != "" {
			return strings.TrimSpace(u)
		}
	}
	return ""
}

func extractFromAny(v any) string {
	switch t := v.(type) {
	case string:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(t)), "data:image/") {
			return strings.TrimSpace(t)
		}
		if m := reMarkdownImage.FindStringSubmatch(t); len(m) > 1 {
			return m[1]
		}
		return rePlainURL.FindString(t)
	case []any:
		for _, it := range t {
			if u := extractFromAny(it); u != "" {
				return u
			}
		}
	case map[string]any:
		for _, k := range []string{"url", "image", "text"} {
			if u := extractFromAny(t[k]); u != "" {
				return u
			}
		}
		if u := extractFromAny(t["image_url"]); u != "" {
			return u
		}
	case map[string]string:
		for _, k := range []string{"url", "image", "text"} {
			if u := extractFromAny(t[k]); u != "" {
				return u
			}
		}
	}
	return ""
}

func (p *Provider) storeDataURL(dataURL string) (string, error) {
	i := strings.Index(dataURL, ",")
	if i <= 0 {
		return "", errors.New("invalid image data url")
	}
	head := strings.ToLower(strings.TrimSpace(dataURL[:i]))
	raw := strings.TrimSpace(dataURL[i+1:])
	if !strings.Contains(head, ";base64") {
		return "", errors.New("image data url is not base64")
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	ext := "png"
	switch {
	case strings.Contains(head, "image/jpeg"):
		ext = "jpg"
	case strings.Contains(head, "image/webp"):
		ext = "webp"
	case strings.Contains(head, "image/gif"):
		ext = "gif"
	}
	outDir := filepath.Join(p.staticRoot, "generated")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("img_%d.%s", time.Now().UnixNano(), ext)
	if err := os.WriteFile(filepath.Join(outDir, name), b, 0o644); err != nil {
		return "", err
	}
	return "/static/generated/" + name, nil
}
