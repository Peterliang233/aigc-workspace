package wuyinkeji

import "testing"

func TestNormalizeModelSegmentAllowsDot(t *testing.T) {
	got, err := normalizeModelSegment("video_veo3.1_fast")
	if err != nil {
		t.Fatalf("normalizeModelSegment returned error: %v", err)
	}
	if got != "video_veo3.1_fast" {
		t.Fatalf("normalizeModelSegment=%q want %q", got, "video_veo3.1_fast")
	}
}

func TestNormalizeModelSegmentRejectsSlash(t *testing.T) {
	if _, err := normalizeModelSegment("api/async/video_veo3.1_fast/extra"); err == nil {
		t.Fatal("normalizeModelSegment should reject multiple path segments")
	}
}
