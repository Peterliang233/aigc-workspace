package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryAnimationSegmentRetriesThenSuccess(t *testing.T) {
	oldDelay := animationSegmentRetryBase
	animationSegmentRetryBase = time.Millisecond
	t.Cleanup(func() { animationSegmentRetryBase = oldDelay })

	attempts := 0
	err := retryAnimationSegment(context.Background(), 3, func(context.Context, int) error {
		attempts++
		if attempts < 3 {
			return errors.New("segment failed")
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

func TestRetryAnimationSegmentStopsOnNonRetryable(t *testing.T) {
	attempts := 0
	err := retryAnimationSegment(context.Background(), 3, func(context.Context, int) error {
		attempts++
		return errors.New("media worker 未配置")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d want 1", attempts)
	}
}

func TestShouldRetryAnimationSegmentErr(t *testing.T) {
	if !shouldRetryAnimationSegmentErr(errors.New("segment generation failed")) {
		t.Fatalf("generic segment failure should be retryable")
	}
	if shouldRetryAnimationSegmentErr(context.Canceled) {
		t.Fatalf("context canceled should not be retryable")
	}
	if shouldRetryAnimationSegmentErr(errors.New("当前 provider 不支持自动生成首帧图，请手动上传参考图片")) {
		t.Fatalf("lead image config errors should not be retryable")
	}
}
