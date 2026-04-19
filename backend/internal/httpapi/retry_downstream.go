package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	downstreamMaxAttempts = 3
)

var downstreamRetryBase = 700 * time.Millisecond

var downstreamStatusPattern = regexp.MustCompile(`\bstatus=(\d{3})\b`)

func retryDownstreamCall(ctx context.Context, op string, fn func(context.Context) error) error {
	if fn == nil {
		return errors.New("retry callback is nil")
	}
	var lastErr error
	for attempt := 1; attempt <= downstreamMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := fn(ctx)
		if err == nil {
			if attempt > 1 {
				slog.Default().Info("downstream_retry_succeeded", "op", op, "attempt", attempt, "max_attempts", downstreamMaxAttempts)
			}
			return nil
		}
		lastErr = err
		if attempt == downstreamMaxAttempts || !shouldRetryDownstreamErr(op, err) {
			break
		}
		wait := time.Duration(attempt) * downstreamRetryBase
		slog.Default().Warn("downstream_retry_scheduled",
			"op", op,
			"attempt", attempt,
			"max_attempts", downstreamMaxAttempts,
			"wait_ms", wait.Milliseconds(),
			"err", err.Error(),
		)
		if !sleepWithContext(ctx, wait) {
			if cerr := ctx.Err(); cerr != nil {
				return cerr
			}
			break
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("downstream call failed")
}

func shouldRetryDownstreamErr(op string, err error) bool {
	if err == nil {
		return false
	}
	if status, ok := extractStatusCode(err); ok {
		if status == 408 || status == 409 || status == 425 || status == 429 || status >= 500 {
			return true
		}
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(msg, "context canceled") || strings.Contains(msg, "context deadline exceeded") {
		return false
	}
	if isCreateDownstreamOp(op) {
		return isSafeCreateRetryErr(err, msg)
	}
	if isRetryableNetErr(err) {
		return true
	}
	for _, marker := range []string{
		"connection reset",
		"connection refused",
		"broken pipe",
		"unexpected eof",
		"tls handshake timeout",
		"i/o timeout",
		"temporarily unavailable",
		"no such host",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"too many requests",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func isCreateDownstreamOp(op string) bool {
	switch strings.TrimSpace(strings.ToLower(op)) {
	case "video_start", "animation_video_start", "image_generate", "audio_generate":
		return true
	default:
		return false
	}
}

func isSafeCreateRetryErr(err error, msg string) bool {
	if isRetryableNetErr(err) {
		if strings.Contains(msg, "tls handshake timeout") {
			return true
		}
		if strings.Contains(msg, "dial tcp") {
			return true
		}
	}
	for _, marker := range []string{
		"tls handshake timeout",
		"no such host",
		"dial tcp",
		"connection refused",
		"server misbehaving",
		"network is unreachable",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func extractStatusCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}
	m := downstreamStatusPattern.FindStringSubmatch(strings.ToLower(err.Error()))
	if len(m) != 2 {
		return 0, false
	}
	code, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		return 0, false
	}
	return code, true
}

func isRetryableNetErr(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary())
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
