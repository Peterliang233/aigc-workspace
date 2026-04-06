package animation

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PromptPlanRequest struct {
	Prompt           string
	TotalSeconds     int
	SegmentDurations []int
}

type plannerSegment struct {
	DurationSeconds int    `json:"duration_seconds"`
	Prompt          string `json:"prompt"`
	Continuity      string `json:"continuity"`
}

type plannerResponse struct {
	Segments []plannerSegment `json:"segments"`
}

func plannerSystemPrompt() string {
	return strings.Join([]string{
		"You are a film previsualization planner for text-to-video generation.",
		"Split one continuous animation idea into multiple seamless segments.",
		"Keep the same subject, outfit, environment, camera direction, lighting, and art style.",
		"Every segment must continue from the exact last frame of the previous segment.",
		"Return only JSON that matches the schema.",
	}, " ")
}

func plannerUserPrompt(req PromptPlanRequest) string {
	payload, _ := json.Marshal(map[string]any{
		"original_prompt":   strings.TrimSpace(req.Prompt),
		"total_duration":    req.TotalSeconds,
		"segment_durations": req.SegmentDurations,
		"requirements": []string{
			"Use the same language as the original prompt.",
			"Generate exactly one segment object for each segment duration, in the same order.",
			"Each prompt must be directly usable by a text-to-video model.",
			"Describe continuous motion and preserve shot continuity.",
			"Do not add numbering, markdown, or explanations inside the prompt text.",
			"Continuity should briefly describe how this segment connects to the previous tail frame.",
		},
	})
	return fmt.Sprintf("Plan the animation segments from this JSON input: %s", string(payload))
}

func plannerSchema() map[string]any {
	return map[string]any{
		"name":   "animation_segment_plan",
		"strict": true,
		"schema": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"segments": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"duration_seconds": map[string]any{"type": "integer"},
							"prompt":           map[string]any{"type": "string"},
							"continuity":       map[string]any{"type": "string"},
						},
						"required": []string{"duration_seconds", "prompt", "continuity"},
					},
				},
			},
			"required": []string{"segments"},
		},
	}
}
