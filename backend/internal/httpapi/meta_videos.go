package httpapi

import (
	"net/http"
	"sort"
	"strings"
)

func (h *Handler) metaVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type prov struct {
		ID         string   `json:"id"`
		Label      string   `json:"label"`
		Configured bool     `json:"configured"`
		Models     []string `json:"models"`
	}

	labels := map[string]string{
		"siliconflow":       "SiliconFlow",
		"openai_compatible": "OpenAI Compatible",
	}

	// Minimal built-in model lists for SiliconFlow video. Users can also type a custom model in UI.
	sfModels := []string{
		"Wan-AI/Wan2.2-T2V-A14B",
		"Wan-AI/Wan2.2-I2V-A14B",
	}

	cfg := h.effectiveCfg()
	var list []prov

	// Only advertise providers that are configured and/or available.
	if pc, ok := cfg.ImageProviders["siliconflow"]; ok {
		configured := strings.TrimSpace(pc.APIKey) != ""
		list = append(list, prov{ID: "siliconflow", Label: labels["siliconflow"], Configured: configured, Models: sfModels})
	}
	if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
		pc := cfg.ImageProviders["openai_compatible"]
		configured := strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" && strings.TrimSpace(cfg.VideoModel) != ""
		models := []string{}
		if strings.TrimSpace(cfg.VideoModel) != "" {
			models = append(models, strings.TrimSpace(cfg.VideoModel))
		}
		list = append(list, prov{ID: "openai_compatible", Label: labels["openai_compatible"], Configured: configured, Models: models})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	def := h.defaultVideoProviderID()

	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": def,
		"providers":        list,
	})
}
