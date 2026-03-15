package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
)

func (h *Handler) metaImages(w http.ResponseWriter, r *http.Request) {
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
		"mock":              "Mock(联调)",
		"openai_compatible": "OpenAI Compatible",
		"siliconflow":       "SiliconFlow",
		"wuyinkeji":         "无印科技(速创API)",
	}

	var list []prov
	cfg := h.effectiveCfg()
	for id, pc := range cfg.ImageProviders {
		if _, ok := labels[id]; !ok {
			continue
		}

		models := pc.Models
		if id == "wuyinkeji" {
			if provImpl, ok := h.getImageProvider("wuyinkeji"); ok {
				// Avoid depending on the concrete provider type: only call Models() when available.
				if ml, ok := provImpl.(interface{ Models() []string }); ok {
					models = ml.Models()
				}
			}
		}

		configured := true
		if id != "mock" && strings.TrimSpace(pc.APIKey) == "" {
			configured = false
		}

		list = append(list, prov{
			ID:         id,
			Label:      labels[id],
			Configured: configured,
			Models:     models,
		})
	}

	// Ensure mock exists even if cfg.ImageProviders was nil/unset.
	foundMock := false
	for _, p := range list {
		if p.ID == "mock" {
			foundMock = true
			break
		}
	}
	if !foundMock {
		list = append(list, prov{ID: "mock", Label: labels["mock"], Configured: true})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": strings.ToLower(strings.TrimSpace(cfg.Provider)),
		"providers":        list,
	})

	slog.Default().Debug("meta_images",
		"default_provider", strings.ToLower(strings.TrimSpace(cfg.Provider)),
		"providers", len(list),
	)
}
