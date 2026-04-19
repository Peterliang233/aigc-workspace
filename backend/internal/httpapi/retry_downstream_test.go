package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestShouldRetryDownstreamErrByStatus(t *testing.T) {
	if !shouldRetryDownstreamErr("video_start", errors.New("video start API error: status=503 body=busy")) {
		t.Fatalf("status=503 should be retryable")
	}
	if !shouldRetryDownstreamErr("video_start", errors.New("video start API error: status=429 body=rate_limited")) {
		t.Fatalf("status=429 should be retryable")
	}
	if shouldRetryDownstreamErr("video_start", errors.New("video start API error: status=400 body=invalid_request")) {
		t.Fatalf("status=400 should not be retryable")
	}
}

func TestRetryDownstreamCallRetriesThenSuccess(t *testing.T) {
	oldDelay := downstreamRetryBase
	downstreamRetryBase = time.Millisecond
	t.Cleanup(func() { downstreamRetryBase = oldDelay })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	attempts := 0
	err := retryDownstreamCall(ctx, "test_retry", func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("video start API error: status=503 body=busy")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d want 3", attempts)
	}
}

func TestRetryDownstreamCallStopsOnNonRetryable(t *testing.T) {
	oldDelay := downstreamRetryBase
	downstreamRetryBase = time.Millisecond
	t.Cleanup(func() { downstreamRetryBase = oldDelay })

	attempts := 0
	err := retryDownstreamCall(context.Background(), "test_no_retry", func(context.Context) error {
		attempts++
		return errors.New("video start API error: status=400 body=invalid_request")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d want 1", attempts)
	}
}

func TestCreateCallRetryOnlyOnSafeNetErr(t *testing.T) {
	if !shouldRetryDownstreamErr("video_start", errors.New(`Post "https://api.bltcy.ai/v2/videos/generations": net/http: TLS handshake timeout`)) {
		t.Fatalf("tls handshake timeout should be retryable for create op")
	}
	if shouldRetryDownstreamErr("video_start", errors.New(`Post "https://api.bltcy.ai/v2/videos/generations": unexpected EOF`)) {
		t.Fatalf("unexpected EOF should not be retried for create op")
	}
	if !shouldRetryDownstreamErr("video_get", errors.New(`Get "https://api.bltcy.ai/v2/videos/generations/id": unexpected EOF`)) {
		t.Fatalf("unexpected EOF should be retryable for get op")
	}
}
