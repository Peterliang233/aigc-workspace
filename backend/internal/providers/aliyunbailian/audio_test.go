package aliyunbailian

import "testing"

func TestGenerationURL(t *testing.T) {
	cases := []struct {
		base string
		want string
	}{
		{
			base: "https://dashscope.aliyuncs.com",
			want: "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation",
		},
		{
			base: "https://dashscope.aliyuncs.com/api/v1",
			want: "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation",
		},
		{
			base: "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation",
			want: "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation",
		},
	}
	for _, tc := range cases {
		got := New(tc.base, "key").generationURL()
		if got != tc.want {
			t.Fatalf("generationURL(%q)=%q, want %q", tc.base, got, tc.want)
		}
	}
}

func TestParseAudioURL(t *testing.T) {
	got, err := parseAudioURL([]byte(`{"output":{"audio":{"url":"https://example.com/a.wav"}},"usage":{"input_tokens":3},"request_id":"rid"}`))
	if err != nil {
		t.Fatalf("parseAudioURL returned error: %v", err)
	}
	if got != "https://example.com/a.wav" {
		t.Fatalf("url=%q", got)
	}
}

func TestParseAudioURLError(t *testing.T) {
	_, err := parseAudioURL([]byte(`{"status_code":400,"code":"InvalidParameter","message":"bad voice"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
