package httpapi

import (
	"strings"

	"aigc-backend/internal/storyvideo"
)

func (h *Handler) defaultStoryPlannerProviderID() string {
	for _, id := range []string{"bltcy", "wuyinkeji", "siliconflow"} {
		if pc, ok := h.cfg.ImageProviders[id]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
			return id
		}
	}
	return "bltcy"
}

func (h *Handler) defaultStoryPlannerModel(providerID string) string {
	switch strings.ToLower(strings.TrimSpace(providerID)) {
	case "wuyinkeji":
		return "gpt-4.1"
	case "siliconflow":
		return "Qwen/Qwen3-32B"
	default:
		return "gpt-5.4"
	}
}

func (h *Handler) storyPlannerFor(providerID string) (*storyvideo.Planner, string, bool) {
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	if providerID == "" {
		providerID = h.defaultStoryPlannerProviderID()
	}
	pc, ok := h.cfg.ImageProviders[providerID]
	if !ok || strings.TrimSpace(pc.BaseURL) == "" || strings.TrimSpace(pc.APIKey) == "" {
		return nil, providerID, false
	}
	return storyvideo.NewPlanner(providerID, pc.BaseURL, pc.APIKey, h.defaultStoryPlannerModel(providerID)), providerID, true
}
