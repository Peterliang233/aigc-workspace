package types

import "strings"

func IsCoreVideoField(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "provider", "model", "prompt", "duration_seconds", "aspect_ratio", "image_size", "negative_prompt", "image", "seed":
		return true
	default:
		return false
	}
}

func (r *VideoJobCreateRequest) SetExtraDefault(key string, value any) {
	key = strings.TrimSpace(key)
	if key == "" || value == nil || IsCoreVideoField(key) {
		return
	}
	if hasVideoValue(r.Extra[key]) {
		return
	}
	if r.Extra == nil {
		r.Extra = map[string]any{}
	}
	r.Extra[key] = value
}

func (r VideoJobCreateRequest) HasFieldValue(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	switch key {
	case "prompt":
		return strings.TrimSpace(r.Prompt) != ""
	case "image":
		return strings.TrimSpace(r.Image) != ""
	case "image_size":
		return strings.TrimSpace(r.ImageSize) != ""
	case "aspect_ratio":
		return strings.TrimSpace(r.AspectRatio) != ""
	case "negative_prompt":
		return strings.TrimSpace(r.NegativePrompt) != ""
	case "duration_seconds":
		return r.DurationSeconds > 0
	case "seed":
		return r.Seed != nil
	default:
		return hasVideoValue(r.Extra[key])
	}
}

func (r VideoJobCreateRequest) MergePayload(base map[string]any) map[string]any {
	if len(r.Extra) == 0 {
		return base
	}
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range r.Extra {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" || IsCoreVideoField(trimmed) || !hasVideoValue(value) {
			continue
		}
		if _, ok := out[trimmed]; ok {
			continue
		}
		out[trimmed] = value
	}
	return out
}

func hasVideoValue(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(x) != ""
	default:
		return true
	}
}
