package storyvideo

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseDraftResponse(raw []byte) (Draft, error) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Draft{}, err
	}
	if len(payload.Choices) == 0 {
		return Draft{}, fmt.Errorf("planner returned empty choices")
	}
	text := extractMessageText(payload.Choices[0].Message.Content)
	if strings.TrimSpace(text) == "" {
		return Draft{}, fmt.Errorf("planner returned empty message content")
	}
	text = unwrapJSONText(text)
	var out Draft
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return Draft{}, err
	}
	return out, nil
}

func extractMessageText(content any) string {
	if s, ok := content.(string); ok {
		return strings.TrimSpace(s)
	}
	items, ok := content.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := entry["text"].(string); ok && strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func unwrapJSONText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```JSON")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}
