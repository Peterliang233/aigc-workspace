package wuyinkeji

import (
	"reflect"
	"testing"

	"aigc-backend/internal/types"
)

func TestBuildStartURLForGPTImage2(t *testing.T) {
	p := New("https://api.wuyinkeji.com", "secret", "/tmp", nil, nil)
	got, err := p.buildStartURL("gpt-image-2")
	if err != nil {
		t.Fatalf("buildStartURL returned error: %v", err)
	}
	want := "https://api.wuyinkeji.com/api/async/image_gpt?key=secret"
	if got != want {
		t.Fatalf("buildStartURL=%q want %q", got, want)
	}
}

func TestBuildImagePayloadForGPTImage2(t *testing.T) {
	payload, err := buildImagePayload("gpt-image-2", types.ImageGenerateRequest{
		Prompt: "draw a cat",
		Size:   "16:9",
		Image:  []string{"https://example.com/a.png", "https://example.com/b.png"},
	})
	if err != nil {
		t.Fatalf("buildImagePayload returned error: %v", err)
	}
	want := map[string]any{
		"prompt": "draw a cat",
		"size":   "16:9",
		"urls":   []string{"https://example.com/a.png", "https://example.com/b.png"},
	}
	if !reflect.DeepEqual(payload, want) {
		t.Fatalf("payload=%#v want %#v", payload, want)
	}
}

func TestBuildImagePayloadForGPTImage2RejectsDataURL(t *testing.T) {
	_, err := buildImagePayload("gpt-image-2", types.ImageGenerateRequest{
		Prompt: "draw a cat",
		Image:  []string{"data:image/png;base64,aaaa"},
	})
	if err == nil {
		t.Fatal("expected error for data url")
	}
}
