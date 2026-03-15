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

	var list []prov
	cfg := h.cfg
	if h.models != nil {
		for _, p := range h.models.Providers {
			if p.Image == nil {
				continue
			}
			id := strings.ToLower(strings.TrimSpace(p.ID))
			if id == "" {
				continue
			}
			configured := true
			if id != "mock" {
				pc := cfg.ImageProviders[id]
				if strings.TrimSpace(pc.APIKey) == "" {
					configured = false
				}
				// Some providers also require a base URL.
				if id == "openai_compatible" && strings.TrimSpace(pc.BaseURL) == "" {
					configured = false
				}
			}
			var models []string
			for _, m := range p.Image.Models {
				mid := strings.TrimSpace(m.ID)
				if mid != "" {
					models = append(models, mid)
				}
			}
			list = append(list, prov{ID: id, Label: p.Label, Configured: configured, Models: models})
		}
	} else {
		// Minimal fallback so UI still renders.
		list = append(list, prov{ID: "mock", Label: "Mock(联调)", Configured: true})
	}

	def := ""
	if h.models != nil {
		def = strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("image")))
	}
	if def == "" {
		def = "mock"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": def,
		"providers":        list,
	})

	slog.Default().Debug("meta_images",
		"default_provider", def,
		"providers", len(list),
	)
}
