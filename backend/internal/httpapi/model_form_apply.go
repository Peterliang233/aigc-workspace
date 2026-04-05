package httpapi

import (
	"strings"

	"aigc-backend/internal/types"
)

func (h *Handler) applyImageModelDefaults(providerID, model string, req *types.ImageGenerateRequest) {
	if h.models == nil {
		return
	}
	ms := h.models.Model(providerID, "image", model)
	if ms == nil || ms.Form == nil {
		return
	}
	for _, f := range ms.Form.Fields {
		k := strings.ToLower(strings.TrimSpace(f.Key))
		if k == "" || f.Default == nil {
			continue
		}
		switch k {
		case "size":
			if strings.TrimSpace(req.Size) == "" {
				if s, ok := asString(f.Default); ok {
					req.Size = s
				}
			}
		case "aspect_ratio":
			if strings.TrimSpace(req.AspectRatio) == "" {
				if s, ok := asString(f.Default); ok {
					req.AspectRatio = s
				}
			}
		case "negative_prompt":
			if strings.TrimSpace(req.NegativePrompt) == "" {
				if s, ok := asString(f.Default); ok {
					req.NegativePrompt = s
				}
			}
		case "seed":
			if req.Seed == nil {
				if n, ok := asInt64(f.Default); ok {
					req.Seed = &n
				}
			}
		case "strength":
			if req.Strength == nil {
				if n, ok := asFloat64(f.Default); ok {
					req.Strength = &n
				}
			}
		case "style":
			if strings.TrimSpace(req.Style) == "" {
				if s, ok := asString(f.Default); ok {
					req.Style = s
				}
			}
		}
	}
}

func (h *Handler) applyVideoModelDefaults(providerID, model string, req *types.VideoJobCreateRequest) {
	if h.models == nil {
		return
	}
	ms := h.models.Model(providerID, "video", model)
	if ms == nil || ms.Form == nil {
		return
	}
	for _, f := range ms.Form.Fields {
		k := strings.ToLower(strings.TrimSpace(f.Key))
		if k == "" || f.Default == nil {
			continue
		}
		switch k {
		case "image_size":
			if strings.TrimSpace(req.ImageSize) == "" {
				if s, ok := asString(f.Default); ok {
					req.ImageSize = s
				}
			}
		case "aspect_ratio":
			if strings.TrimSpace(req.AspectRatio) == "" {
				if s, ok := asString(f.Default); ok {
					req.AspectRatio = s
				}
			}
		case "negative_prompt":
			if strings.TrimSpace(req.NegativePrompt) == "" {
				if s, ok := asString(f.Default); ok {
					req.NegativePrompt = s
				}
			}
		case "duration_seconds":
			if req.DurationSeconds == 0 {
				if n, ok := asInt64(f.Default); ok && n > 0 {
					req.DurationSeconds = int(n)
				}
			}
		case "seed":
			if req.Seed == nil {
				if n, ok := asInt64(f.Default); ok {
					req.Seed = &n
				}
			}
		default:
			req.SetExtraDefault(k, f.Default)
		}
	}
}
