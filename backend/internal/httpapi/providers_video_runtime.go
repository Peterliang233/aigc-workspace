package httpapi

import (
	"strings"
)

func (h *Handler) defaultVideoProviderID() string {
	cfg := h.effectiveCfg()
	if pc, ok := cfg.ImageProviders["siliconflow"]; ok && strings.TrimSpace(pc.APIKey) != "" {
		return "siliconflow"
	}
	if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
		pc := cfg.ImageProviders["openai_compatible"]
		if strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
			return "openai_compatible"
		}
	}
	return ""
}

func (h *Handler) getVideoProvider(id string) (videoProvider, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		id = h.defaultVideoProviderID()
	}
	if id == "" {
		return nil, false
	}

	h.provMu.Lock()
	defer h.provMu.Unlock()
	h.rebuildProvidersLocked()
	p, ok := h.videoProviders[id]
	return p, ok
}
