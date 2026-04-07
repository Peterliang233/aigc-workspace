package logging

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestDownstreamBodyBytesDisabled(t *testing.T) {
	t.Setenv("LOG_DOWNSTREAM_RAW_BODY", "false")
	got, encoding := downstreamBodyBytes("application/json", []byte(`{"ok":true}`))
	if got != "" || encoding != "" {
		t.Fatalf("got (%q, %q), want empty", got, encoding)
	}
}

func TestDownstreamBodyBytesText(t *testing.T) {
	t.Setenv("LOG_DOWNSTREAM_RAW_BODY", "true")
	t.Setenv("LOG_DOWNSTREAM_BODY_FULL", "true")
	got, encoding := downstreamBodyBytes("application/json", []byte(`{"ok":true}`))
	if got != `{"ok":true}` {
		t.Fatalf("got %q", got)
	}
	if encoding != "text" {
		t.Fatalf("encoding = %q, want text", encoding)
	}
}

func TestDownstreamBodyBytesBinary(t *testing.T) {
	t.Setenv("LOG_DOWNSTREAM_RAW_BODY", "true")
	t.Setenv("LOG_DOWNSTREAM_BODY_FULL", "true")
	payload := []byte{0xff, 0xd8, 0xff, 0x00}
	got, encoding := downstreamBodyBytes("audio/mpeg", payload)
	if got != base64.StdEncoding.EncodeToString(payload) {
		t.Fatalf("got %q", got)
	}
	if encoding != "base64" {
		t.Fatalf("encoding = %q, want base64", encoding)
	}
}

func TestDownstreamBodyBytesTruncates(t *testing.T) {
	t.Setenv("LOG_DOWNSTREAM_RAW_BODY", "true")
	_ = os.Setenv("LOG_DOWNSTREAM_BODY_FULL", "false")
	t.Setenv("LOG_DOWNSTREAM_BODY_CHARS", "4")
	got, _ := downstreamBodyBytes("application/json", []byte(`{"ok":true}`))
	if got != `{"ok...` {
		t.Fatalf("got %q", got)
	}
}
