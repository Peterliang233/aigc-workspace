package httpapi

import (
	"net/http"
	"strings"

	"aigc-backend/internal/providers/wuyinkeji"
)

func (h *Handler) settingsGet(w http.ResponseWriter, r *http.Request) {
	type prov struct {
		Label        string   `json:"label"`
		BaseURL      string   `json:"base_url,omitempty"`
		APIKeySet    bool     `json:"api_key_set"`
		DefaultModel string   `json:"default_model,omitempty"`
		Models       []string `json:"models,omitempty"`
	}

	labels := map[string]string{
		"openai_compatible": "OpenAI Compatible",
		"siliconflow":       "SiliconFlow",
		"wuyinkeji":         "无印科技(速创API)",
	}

	cfg := h.effectiveCfg()
	out := map[string]prov{}
	for id, label := range labels {
		pc := cfg.ImageProviders[id]
		models := pc.Models
		if id == "wuyinkeji" {
			if provImpl, ok := h.getImageProvider("wuyinkeji"); ok {
				if p, ok := provImpl.(*wuyinkeji.Provider); ok {
					models = p.Models()
				}
			}
		}

		out[id] = prov{
			Label:        label,
			BaseURL:      strings.TrimSpace(pc.BaseURL),
			APIKeySet:    strings.TrimSpace(pc.APIKey) != "",
			DefaultModel: strings.TrimSpace(pc.DefaultModel),
			Models:       models,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"image_providers": out,
	})
}

