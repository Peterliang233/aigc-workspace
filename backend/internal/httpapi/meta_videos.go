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

	type model struct {
		ID            string `json:"id"`
		Label         string `json:"label,omitempty"`
		RequiresImage bool   `json:"requires_image,omitempty"`
	}
	type prov struct {
		ID         string  `json:"id"`
		Label      string  `json:"label"`
		Configured bool    `json:"configured"`
		Models     []model `json:"models"`
	}

	cfg := h.cfg
	var list []prov

	if h.models != nil {
		for _, p := range h.models.Providers {
			if p.Video == nil {
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
				if id == "openai_compatible" && strings.TrimSpace(pc.BaseURL) == "" {
					configured = false
				}
			}
			// openai_compatible videos need extra env endpoints.
			if id == "openai_compatible" && (cfg.VideoStartEP == "" || cfg.VideoStatusEP == "") {
				configured = false
			}

			var ms []model
			for _, m := range p.Video.Models {
				mid := strings.TrimSpace(m.ID)
				if mid == "" {
					continue
				}
				reqImg := false
				if m.Form != nil {
					reqImg = m.Form.RequiresImage
				}
				ms = append(ms, model{ID: mid, Label: strings.TrimSpace(m.Label), RequiresImage: reqImg})
			}
			list = append(list, prov{ID: id, Label: p.Label, Configured: configured, Models: ms})
		}
	} else {
		list = append(list, prov{ID: "siliconflow", Label: "SiliconFlow", Configured: false, Models: nil})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	def := ""
	if h.models != nil {
		def = strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("video")))
	}
	if def == "" {
		def = h.defaultVideoProviderID()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": def,
		"providers":        list,
	})
}
