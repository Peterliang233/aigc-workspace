package wuyinkeji

import (
	"errors"
	"regexp"
	"strings"
)

func mapSize(size string) string {
	// doc accepts: 1K,2K,4K
	s := strings.TrimSpace(strings.ToLower(size))
	switch s {
	case "1024x1024", "1k":
		return "1K"
	case "2048x2048", "2k":
		return "2K"
	case "4096x4096", "4k":
		return "4K"
	default:
		return "1K"
	}
}

func mapAspect(ar string) string {
	ar = strings.TrimSpace(ar)
	if ar == "" {
		return "1:1"
	}
	// allow common ratios used in our app
	switch ar {
	case "1:1", "16:9", "9:16", "4:3", "3:4", "2:3", "3:2":
		return ar
	default:
		return "1:1"
	}
}

var modelSegRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func normalizeModelSegment(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("model is required")
	}
	s = strings.TrimPrefix(s, "/api/async/")
	s = strings.TrimPrefix(s, "api/async/")
	s = strings.TrimPrefix(s, "/")
	// Only allow a single safe path segment to avoid SSRF/path traversal.
	if strings.Contains(s, "/") {
		return "", errors.New("invalid model name")
	}
	if !modelSegRe.MatchString(s) {
		return "", errors.New("invalid model name")
	}
	return s, nil
}
