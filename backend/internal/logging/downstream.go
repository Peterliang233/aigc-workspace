package logging

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
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
	if strings.TrimSpace(event) == "" {
		event = "downstream_request"
	}
	slog.Default().Info(event,
		"provider", strings.TrimSpace(provider),
		"method", strings.TrimSpace(method),
		"url", RedactURL(rawURL),
		"params", params,
	)
}

func DownstreamRequestDebug(event, provider, method, rawURL string, params any) {
	if strings.TrimSpace(event) == "" {
		event = "downstream_request"
	}
	slog.Default().Debug(event,
		"provider", strings.TrimSpace(provider),
		"method", strings.TrimSpace(method),
		"url", RedactURL(rawURL),
		"params", params,
	)
}

// DownstreamResponse logs an outbound response.
func DownstreamResponse(event, provider, method, rawURL string, status int, dur time.Duration, err error) {
	if strings.TrimSpace(event) == "" {
		event = "downstream_response"
	}
	fields := []any{
		"provider", strings.TrimSpace(provider),
		"method", strings.TrimSpace(method),
		"url", RedactURL(rawURL),
		"status", status,
		"dur_ms", dur.Milliseconds(),
	}
	if err != nil {
		fields = append(fields, "err", err.Error())
		slog.Default().Warn(event, fields...)
		return
	}
	slog.Default().Info(event, fields...)
}

func DownstreamResponseDebug(event, provider, method, rawURL string, status int, dur time.Duration, err error) {
	if strings.TrimSpace(event) == "" {
		event = "downstream_response"
	}
	fields := []any{
		"provider", strings.TrimSpace(provider),
		"method", strings.TrimSpace(method),
		"url", RedactURL(rawURL),
		"status", status,
		"dur_ms", dur.Milliseconds(),
	}
	if err != nil {
		fields = append(fields, "err", err.Error())
		slog.Default().Debug(event, fields...)
		return
	}
	slog.Default().Debug(event, fields...)
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
