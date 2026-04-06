package gptbest

import "testing"

func TestApplyGptBestVideoImage(t *testing.T) {
	body := map[string]any{"model": "wan2.2-i2v-plus"}
	applyGptBestVideoImage(body, "https://example.com/input.png")
	if got := body["image"]; got != "https://example.com/input.png" {
		t.Fatalf("image mismatch: %#v", got)
	}
	if got := body["img_url"]; got != "https://example.com/input.png" {
		t.Fatalf("img_url mismatch: %#v", got)
	}
	images, ok := body["images"].([]string)
	if !ok || len(images) != 1 || images[0] != "https://example.com/input.png" {
		t.Fatalf("images mismatch: %#v", body["images"])
	}
}

func TestFirstNonEmptyString(t *testing.T) {
	if got := firstNonEmptyString("", "  ", "x"); got != "x" {
		t.Fatalf("unexpected first non empty string: %q", got)
	}
}
