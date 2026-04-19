package wuyinkeji

import (
	"testing"

	"aigc-backend/internal/types"
)

func TestBuildVideoPayloadUsesVeoURLsField(t *testing.T) {
	payload := buildVideoPayload("video_veo3.1_fast", types.VideoJobCreateRequest{
		Prompt: "test",
		Extra: map[string]any{
			"reference_url": "https://example.com/ref.png",
		},
	})
	if _, ok := payload["urls"]; !ok {
		t.Fatal("expected veo payload to use urls field")
	}
	if _, ok := payload["image_urls"]; ok {
		t.Fatal("did not expect veo payload to use image_urls field")
	}
}

func TestBuildVideoPayloadKeepsLegacyImageURLsField(t *testing.T) {
	payload := buildVideoPayload("video_grok_imagine", types.VideoJobCreateRequest{
		Prompt: "test",
		Extra: map[string]any{
			"reference_url": "https://example.com/ref.png",
		},
	})
	if _, ok := payload["image_urls"]; !ok {
		t.Fatal("expected non-veo payload to use image_urls field")
	}
	if _, ok := payload["urls"]; ok {
		t.Fatal("did not expect non-veo payload to use urls field")
	}
}
