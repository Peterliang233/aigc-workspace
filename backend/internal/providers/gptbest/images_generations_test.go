package gptbest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aigc-backend/internal/types"
)

func TestGenerateImageUsesImagesAPIForGptImageModels(t *testing.T) {
	tmp := t.TempDir()
	var path string
	p := New("bltcy", "https://example.com", "test-key", "gpt-image-2", tmp)
	p.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		path = r.URL.Path
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("path = %q, want /v1/images/generations", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got := body["model"]; got != "gpt-image-2" {
			t.Fatalf("model = %#v, want gpt-image-2", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"` + base64.StdEncoding.EncodeToString([]byte("png")) + `"}]}`)),
		}, nil
	})}

	out, err := p.GenerateImage(context.Background(), types.ImageGenerateRequest{
		Prompt: "draw a cat",
		Model:  "gpt-image-2",
	})
	if err != nil {
		t.Fatalf("GenerateImage error: %v", err)
	}
	if path != "/v1/images/generations" {
		t.Fatalf("request path = %q", path)
	}
	if len(out.ImageURLs) != 1 {
		t.Fatalf("image urls = %#v", out.ImageURLs)
	}
	if got := out.ImageURLs[0]; filepath.Dir(got) != "/static/generated" {
		t.Fatalf("image url = %q", got)
	}
	if _, err := os.Stat(filepath.Join(tmp, "generated", filepath.Base(out.ImageURLs[0]))); err != nil {
		t.Fatalf("stored file missing: %v", err)
	}
}

func TestPrefersImagesAPIModel(t *testing.T) {
	if !prefersImagesAPIModel("gpt-image-2") {
		t.Fatal("gpt-image-2 should use images api")
	}
	if prefersImagesAPIModel("qwen-image") {
		t.Fatal("qwen-image should not use images api")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
