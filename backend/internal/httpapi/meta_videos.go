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

	type fieldOption struct {
		Label string `json:"label,omitempty"`
		Value string `json:"value"`
	}
	type field struct {
		Key         string        `json:"key"`
		Label       string        `json:"label,omitempty"`
		Type        string        `json:"type"`
		Required    bool          `json:"required,omitempty"`
		Placeholder string        `json:"placeholder,omitempty"`
		Default     any           `json:"default,omitempty"`
		Options     []fieldOption `json:"options,omitempty"`
		Rows        int           `json:"rows,omitempty"`
	}
	type form struct {
		RequiresImage bool    `json:"requires_image,omitempty"`
		Fields        []field `json:"fields,omitempty"`
	}
	type model struct {
		ID            string `json:"id"`
		Label         string `json:"label,omitempty"`
		RequiresImage bool   `json:"requires_image,omitempty"`
		Form          *form  `json:"form,omitempty"`
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
				var outForm *form
				if m.Form != nil {
					reqImg = m.Form.RequiresImage
					outForm = &form{RequiresImage: m.Form.RequiresImage}
					for _, f := range m.Form.Fields {
						of := field{
							Key:         strings.TrimSpace(f.Key),
							Label:       strings.TrimSpace(f.Label),
							Type:        strings.TrimSpace(f.Type),
							Required:    f.Required,
							Placeholder: strings.TrimSpace(f.Placeholder),
							Default:     f.Default,
							Rows:        f.Rows,
						}
						for _, opt := range f.Options {
							of.Options = append(of.Options, fieldOption{
								Label: strings.TrimSpace(opt.Label),
								Value: strings.TrimSpace(opt.Value),
							})
						}
						if of.Key != "" && of.Type != "" {
							outForm.Fields = append(outForm.Fields, of)
						}
						if strings.EqualFold(of.Key, "image") && of.Required {
							reqImg = true
						}
					}
				}
				ms = append(ms, model{ID: mid, Label: strings.TrimSpace(m.Label), RequiresImage: reqImg, Form: outForm})
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
