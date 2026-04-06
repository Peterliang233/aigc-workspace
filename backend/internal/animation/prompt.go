package animation

import (
	"fmt"
	"strings"
)

func BuildSegmentPrompt(prompt string, index, total int) string {
	stage := segmentStage(index, total)
	parts := []string{
		strings.TrimSpace(prompt),
		fmt.Sprintf("Segment %d of %d.", index+1, total),
		stage,
		"Keep the same subject, outfit, environment, camera direction, lighting, and visual style.",
		"Continue motion naturally from the previous segment without a hard cut.",
	}
	return strings.Join(parts, " ")
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
