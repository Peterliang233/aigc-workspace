package logging

import (
	"encoding/base64"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"
)

func DownstreamRequestRaw(event, provider, method, rawURL string, params any, contentType string, body []byte) {
	if strings.TrimSpace(event) == "" {
		event = "downstream_request"
	}
	fields := []any{
		"provider", strings.TrimSpace(provider),
		"method", strings.TrimSpace(method),
		"url", RedactURL(rawURL),
		"params", params,
	}
	slog.Default().Info(event, appendBody(fields, contentType, body)...)
}

func DownstreamRequestDebugRaw(event, provider, method, rawURL string, params any, contentType string, body []byte) {
	if strings.TrimSpace(event) == "" {
		event = "downstream_request"
	}
	fields := []any{
		"provider", strings.TrimSpace(provider),
		"method", strings.TrimSpace(method),
		"url", RedactURL(rawURL),
		"params", params,
	}
	slog.Default().Debug(event, appendBody(fields, contentType, body)...)
}

func DownstreamResponseRaw(event, provider, method, rawURL string, status int, dur time.Duration, err error, contentType string, body []byte) {
	logDownstreamResponse(slog.LevelInfo, event, provider, method, rawURL, status, dur, err, contentType, body)
}

func DownstreamResponseDebugRaw(event, provider, method, rawURL string, status int, dur time.Duration, err error, contentType string, body []byte) {
	logDownstreamResponse(slog.LevelDebug, event, provider, method, rawURL, status, dur, err, contentType, body)
}

func logDownstreamResponse(level slog.Level, event, provider, method, rawURL string, status int, dur time.Duration, err error, contentType string, body []byte) {
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
	fields = appendBody(fields, contentType, body)
	if err != nil {
		fields = append(fields, "err", err.Error())
	}
	if level == slog.LevelDebug {
		slog.Default().Debug(event, fields...)
		return
	}
	if err != nil {
		slog.Default().Warn(event, fields...)
		return
	}
	slog.Default().Info(event, fields...)
}

func appendBody(fields []any, contentType string, body []byte) []any {
	text, encoding := downstreamBodyBytes(contentType, body)
	if text == "" {
		return fields
	}
	if ct := strings.TrimSpace(contentType); ct != "" {
		fields = append(fields, "content_type", ct)
	}
	if encoding != "" && encoding != "text" {
		fields = append(fields, "body_encoding", encoding)
	}
	return append(fields, "body", text)
}

func downstreamBodyBytes(contentType string, body []byte) (string, string) {
	if !envBool("LOG_DOWNSTREAM_RAW_BODY") || len(body) == 0 {
		return "", ""
	}
	if isTextBody(contentType, body) {
		return downstreamBody(string(body)), "text"
	}
	return downstreamBody(base64.StdEncoding.EncodeToString(body)), "base64"
}

func isTextBody(contentType string, body []byte) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(ct, "text/"),
		strings.Contains(ct, "json"),
		strings.Contains(ct, "xml"),
		strings.Contains(ct, "javascript"),
		strings.Contains(ct, "form-urlencoded"):
		return true
	}
	return utf8.Valid(body)
}
