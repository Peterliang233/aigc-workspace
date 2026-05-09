package gptbest

import "testing"

func TestStoreBase64ImagePassesThroughURL(t *testing.T) {
	p := New("bltcy", "https://example.com", "key", "gpt-image-2", t.TempDir())
	got, err := p.storeBase64Image("https://example.com/image.png", "png")
	if err != nil {
		t.Fatalf("storeBase64Image returned error: %v", err)
	}
	if got != "https://example.com/image.png" {
		t.Fatalf("storeBase64Image=%q want url passthrough", got)
	}
}

func TestStoreBase64ImageAcceptsDataURL(t *testing.T) {
	p := New("bltcy", "https://example.com", "key", "gpt-image-2", t.TempDir())
	got, err := p.storeBase64Image("data:image/png;base64,cG5n", "png")
	if err != nil {
		t.Fatalf("storeBase64Image returned error: %v", err)
	}
	if got == "" {
		t.Fatal("storeBase64Image returned empty path")
	}
}
