package httpapi

import "aigc-backend/internal/storyvideo"

func storyVideoOriginalDurations(shots []storyvideo.Shot) []int {
	out := make([]int, len(shots))
	for i, shot := range shots {
		out[i] = maxInt(1000, shot.DurationMS)
	}
	return out
}
