package gptbest

import "strings"

func applyGptBestVideoImage(body map[string]any, image string) {
	image = strings.TrimSpace(image)
	if image == "" {
		return
	}
	body["image"] = image
	body["img_url"] = image
	if _, ok := body["images"]; !ok {
		body["images"] = []string{image}
	}
}

func firstNonEmptyString(values ...any) string {
	for _, value := range values {
		s := strings.TrimSpace(scalarToString(value))
		if s != "" {
			return s
		}
	}
	return ""
}
