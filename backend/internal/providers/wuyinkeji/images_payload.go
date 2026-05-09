package wuyinkeji

import (
	"errors"
	"strings"

	"aigc-backend/internal/types"
)

const imageGPTEndpoint = "/api/async/image_gpt"

func buildImagePayload(model string, req types.ImageGenerateRequest) (map[string]any, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if isGPTImage2Model(model) {
		payload := map[string]any{
			"prompt": prompt,
			"size":   mapGPTImage2Size(req.Size, req.AspectRatio),
		}
		urls, err := publicImageURLs(req.Image, req.ReferenceURLs)
		if err != nil {
			return nil, err
		}
		if len(urls) > 0 {
			payload["urls"] = urls
		}
		return payload, nil
	}
	return map[string]any{
		"prompt":      prompt,
		"size":        mapSize(req.Size),
		"aspectRatio": mapAspect(req.AspectRatio),
	}, nil
}

func isGPTImage2Model(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "gpt-image-2")
}

func mapGPTImage2Size(size, aspect string) string {
	value := strings.TrimSpace(size)
	if value == "" {
		value = strings.TrimSpace(aspect)
	}
	switch value {
	case "", "auto":
		return "auto"
	case "1:1", "3:2", "2:3", "16:9", "9:16", "4:3", "3:4", "21:9", "9:21", "1:3", "3:1", "2:1", "1:2":
		return value
	default:
		return "auto"
	}
}

func publicImageURLs(groups ...[]string) ([]string, error) {
	out := make([]string, 0, 2)
	for _, group := range groups {
		for _, ref := range group {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}
			if strings.HasPrefix(strings.ToLower(ref), "data:") {
				return nil, errors.New("速创 GPT-Image-2 的参考图片仅支持公网 URL，不支持本地上传图片")
			}
			if strings.HasPrefix(strings.ToLower(ref), "http://") || strings.HasPrefix(strings.ToLower(ref), "https://") {
				out = append(out, ref)
			}
		}
	}
	return out, nil
}
