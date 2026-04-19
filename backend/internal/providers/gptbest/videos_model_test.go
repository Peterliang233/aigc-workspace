package gptbest

import "testing"

func TestNormalizeVideoModel(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		model string
	}{
		{name: "keep normal model", raw: "wan2.2-t2v-plus", model: "wan2.2-t2v-plus"},
		{name: "legacy veo underscore", raw: "veo3.1_fast", model: "veo3.1-fast"},
		{name: "legacy prefixed veo", raw: "video_veo3.1_fast", model: "veo3.1-fast"},
		{name: "blank", raw: "   ", model: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalizeVideoModel(test.raw)
			if got != test.model {
				t.Fatalf("normalizeVideoModel(%q)=%q want %q", test.raw, got, test.model)
			}
		})
	}
}
