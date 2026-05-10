package textgen

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseResponse(raw []byte) (string, error) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	if len(payload.Choices) == 0 {
		return "", fmt.Errorf("text provider returned empty choices")
	}
	text := extractText(payload.Choices[0].Message.Content)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("text provider returned empty content")
	}
	return strings.TrimSpace(text), nil
}

func extractText(content any) string {
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
