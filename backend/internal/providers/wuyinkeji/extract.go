package wuyinkeji

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

var urlRe = regexp.MustCompile(`https?://[^\\s\"']+`)

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func extractURLs(v any) []string {
	// Robust fallback: scan for http(s) strings in the JSON-encoded data.
	b, _ := json.Marshal(v)
	m := urlRe.FindAllString(string(b), -1)
	seen := map[string]bool{}
	var out []string
	for _, u := range m {
		u = strings.Trim(u, `\"`)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

func extractStatus(v any) (status int, message string) {
	// Expected fields from docs:
	// data.status: 0 queued, 1 running, 2 success, 3 failed
	// data.message: error message
	m, ok := v.(map[string]any)
	if !ok {
		return 0, ""
	}
	if raw, ok := m["status"]; ok {
		switch t := raw.(type) {
		case float64:
			status = int(t)
		case int:
			status = t
		case string:
			// ignore
		}
	}
	if raw, ok := m["message"]; ok {
		if s, ok := raw.(string); ok {
			message = s
		}
	}
	return status, message
}

func ioReadAllLimit(r io.Reader, n int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, n))
}

