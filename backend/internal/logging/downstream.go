package logging

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"
)

// DownstreamPrompt returns a safe-to-log representation of a prompt.
//
// Defaults:
// - includes len + sha256
// - includes a preview (first 120 runes) unless LOG_PROMPT_PREVIEW_CHARS=0
// - full prompt only when LOG_PROMPT_FULL=true
func DownstreamPrompt(prompt string) any {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(prompt))
	out := map[string]any{
		"len":    len([]rune(prompt)),
		"sha256": hex.EncodeToString(sum[:]),
	}

	if envBool("LOG_PROMPT_FULL") {
		out["full"] = prompt
		return out
	}

	n := envInt("LOG_PROMPT_PREVIEW_CHARS", 120)
	if n <= 0 {
		return out
	}

	prev, truncated := truncRunes(prompt, n)
	if truncated {
		prev = prev + "..."
	}
	out["preview"] = prev
	return out
}

// DownstreamRequest logs an outbound request to a vendor/provider.
// URL is always redacted (drops query/fragment).
func DownstreamRequest(event, provider, method, rawURL string, params any) {
	DownstreamRequestRaw(event, provider, method, rawURL, params, "", nil)
}

func DownstreamRequestDebug(event, provider, method, rawURL string, params any) {
	DownstreamRequestDebugRaw(event, provider, method, rawURL, params, "", nil)
}

// DownstreamResponse logs an outbound response.
func DownstreamResponse(event, provider, method, rawURL string, status int, dur time.Duration, err error, body ...string) {
	DownstreamResponseRaw(event, provider, method, rawURL, status, dur, err, "", firstBodyBytes(body...))
}

func DownstreamResponseDebug(event, provider, method, rawURL string, status int, dur time.Duration, err error, body ...string) {
	DownstreamResponseDebugRaw(event, provider, method, rawURL, status, dur, err, "", firstBodyBytes(body...))
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func truncRunes(s string, n int) (string, bool) {
	if n <= 0 {
		return "", len(s) > 0
	}
	r := []rune(s)
	if len(r) <= n {
		return s, false
	}
	return string(r[:n]), true
}

func downstreamBody(body ...string) string {
	if len(body) == 0 {
		return ""
	}
	text := strings.TrimSpace(body[0])
	if text == "" {
		return ""
	}
	limit := envInt("LOG_DOWNSTREAM_BODY_CHARS", 1500)
	if envBool("LOG_DOWNSTREAM_BODY_FULL") || limit <= 0 {
		return text
	}
	out, truncated := truncRunes(text, limit)
	if truncated {
		return out + "..."
	}
	return out
}

func firstBodyBytes(body ...string) []byte {
	if len(body) == 0 || strings.TrimSpace(body[0]) == "" {
		return nil
	}
	return []byte(body[0])
}
