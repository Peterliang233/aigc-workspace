package httpapi

import (
	"testing"

	"aigc-backend/internal/storyvideo"
)

func TestStoryVideoNarrationForAudioUsesShotLines(t *testing.T) {
	shots := []storyvideo.Shot{{NarrationLine: "第一幕。"}, {NarrationLine: "第二幕。"}}
	got := storyVideoNarrationForAudio("旧的整段解说", shots)
	want := "第一幕。\n第二幕。"
	if got != want {
		t.Fatalf("narration=%q want=%q", got, want)
	}
}

func TestStoryVideoOriginalDurations(t *testing.T) {
	shots := []storyvideo.Shot{
		{DurationMS: 800},
		{DurationMS: 3200},
	}
	got := storyVideoOriginalDurations(shots)
	want := []int{1000, 3200}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%d want=%d", i, got[i], want[i])
		}
	}
}
