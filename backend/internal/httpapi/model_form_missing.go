package httpapi

import (
	"strings"

	"aigc-backend/internal/types"
)

func (h *Handler) missingImageRequiredFields(providerID, model string, req types.ImageGenerateRequest) []string {
	if h.models == nil {
		return nil
	}
	ms := h.models.Model(providerID, "image", model)
	if ms == nil || ms.Form == nil {
		return nil
	}
	var miss []string
	for _, f := range ms.Form.Fields {
		if !f.Required {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(f.Key))
		switch k {
		case "prompt":
			if strings.TrimSpace(req.Prompt) == "" {
				miss = append(miss, "prompt")
			}
		case "size":
			if strings.TrimSpace(req.Size) == "" {
				miss = append(miss, "size")
			}
		case "aspect_ratio":
			if strings.TrimSpace(req.AspectRatio) == "" {
				miss = append(miss, "aspect_ratio")
			}
		case "seed":
			if req.Seed == nil {
				miss = append(miss, "seed")
			}
		case "image":
			if len(cleanImageRefs(req.Image, req.ReferenceURLs)) == 0 {
				miss = append(miss, "image")
			}
		case "reference_urls":
			if len(cleanImageRefs(req.ReferenceURLs, req.Image)) == 0 {
				miss = append(miss, "reference_urls")
			}
		case "strength":
			if req.Strength == nil {
				miss = append(miss, "strength")
			}
		}
	}
	return miss
}

func cleanImageRefs(groups ...[]string) []string {
	out := make([]string, 0, 2)
	for _, g := range groups {
		for _, s := range g {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func (h *Handler) missingVideoRequiredFields(providerID, model string, req types.VideoJobCreateRequest) []string {
	if h.models == nil {
		return nil
	}
	ms := h.models.Model(providerID, "video", model)
	if ms == nil || ms.Form == nil {
		return nil
	}
	var miss []string
	for _, f := range ms.Form.Fields {
		if !f.Required {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(f.Key))
		switch k {
		case "prompt":
			if strings.TrimSpace(req.Prompt) == "" {
				miss = append(miss, "prompt")
			}
		case "image":
			if strings.TrimSpace(req.Image) == "" {
				miss = append(miss, "image")
			}
		case "image_size":
			if strings.TrimSpace(req.ImageSize) == "" {
				miss = append(miss, "image_size")
			}
		case "aspect_ratio":
			if strings.TrimSpace(req.AspectRatio) == "" {
				miss = append(miss, "aspect_ratio")
			}
		case "duration_seconds":
			if req.DurationSeconds <= 0 {
				miss = append(miss, "duration_seconds")
			}
		default:
			if !req.HasFieldValue(k) {
				miss = append(miss, k)
			}
		}
	}

	// keep legacy flag too
	if h.modelRequiresInitImage(providerID, model) && strings.TrimSpace(req.Image) == "" && !containsStr(miss, "image") {
		miss = append(miss, "image")
	}
	return miss
}
