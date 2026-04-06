package animation

import (
	"fmt"
	"strings"
)

type PlannedSegment struct {
	DurationSeconds int
	Prompt          string
	Continuity      string
}

func BuildSegmentPrompt(prompt string, index, total int) string {
	stage := segmentStage(index, total)
	parts := []string{
		strings.TrimSpace(prompt),
		fmt.Sprintf("Segment %d of %d.", index+1, total),
		stage,
		"Keep the same subject, outfit, environment, camera direction, lighting, and visual style.",
		segmentContinuity(index, total),
	}
	return strings.Join(parts, " ")
}

func BuildFallbackSegments(prompt string, plan []int) []PlannedSegment {
	out := make([]PlannedSegment, 0, len(plan))
	for idx, dur := range plan {
		out = append(out, PlannedSegment{
			DurationSeconds: dur,
			Prompt:          BuildSegmentPrompt(prompt, idx, len(plan)),
			Continuity:      segmentContinuity(idx, len(plan)),
		})
	}
	return out
}

func segmentStage(index, total int) string {
	switch {
	case index == 0:
		return "Establish the scene and begin with smooth motion."
	case index == total-1:
		return "Continue the same motion and close naturally without changing the shot."
	default:
		return "Continue the same shot with smooth camera and subject motion."
	}
}

func segmentContinuity(index, total int) string {
	if index == 0 {
		return "Start a single continuous shot that can be extended by later segments."
	}
	if index == total-1 {
		return "Match the previous tail frame exactly and finish the same shot naturally without a hard cut."
	}
	return "Match the previous tail frame exactly and continue the same shot without a hard cut."
}
