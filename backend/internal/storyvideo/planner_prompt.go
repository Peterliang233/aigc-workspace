package storyvideo

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DraftRequest struct {
	Keywords        []string
	Theme           string
	Audience        string
	Tone            string
	DurationSeconds int
	AspectRatio     string
	Extra           string
}

type DraftShot struct {
	Title         string `json:"title"`
	StoryBeat     string `json:"story_beat"`
	NarrationLine string `json:"narration_line"`
	ImagePrompt   string `json:"image_prompt"`
	DurationMS    int    `json:"duration_ms"`
}

type Draft struct {
	Title         string      `json:"title"`
	Summary       string      `json:"summary"`
	ScriptText    string      `json:"script_text"`
	NarrationText string      `json:"narration_text"`
	Shots         []DraftShot `json:"shots"`
}

func plannerSystemPrompt() string {
	return strings.Join([]string{
		"You are a story video planner for narrated slideshow videos.",
		"Turn keywords into a short story outline, narration, and storyboard shots.",
		"Every string field must be concrete and non-empty.",
		"Every shot must include a non-empty title, story_beat, narration_line, image_prompt, and positive duration_ms.",
		"Return only valid JSON, with no markdown fences and no explanation.",
		`Output shape: {"title":"","summary":"","script_text":"","narration_text":"","shots":[{"title":"","story_beat":"","narration_line":"","image_prompt":"","duration_ms":4000}]}`,
	}, " ")
}

func plannerUserPrompt(req DraftRequest) string {
	payload, _ := json.Marshal(map[string]any{
		"keywords":         req.Keywords,
		"theme":            strings.TrimSpace(req.Theme),
		"audience":         strings.TrimSpace(req.Audience),
		"tone":             strings.TrimSpace(req.Tone),
		"duration_seconds": req.DurationSeconds,
		"aspect_ratio":     strings.TrimSpace(req.AspectRatio),
		"extra":            strings.TrimSpace(req.Extra),
	})
	return fmt.Sprintf("Create a narrated storyboard from: %s", string(payload))
}
