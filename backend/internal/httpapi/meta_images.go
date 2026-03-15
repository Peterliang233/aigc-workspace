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
		ID    string `json:"id"`
		Label string `json:"label,omitempty"`
		Form  *form  `json:"form,omitempty"`
	}
	type prov struct {
		ID         string  `json:"id"`
		Label      string  `json:"label"`
		Configured bool    `json:"configured"`
		Models     []model `json:"models"`
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
			var models []model
			for _, m := range p.Image.Models {
				mid := strings.TrimSpace(m.ID)
				if mid == "" {
					continue
				}
				var outForm *form
				if m.Form != nil {
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
					}
				}
				models = append(models, model{ID: mid, Label: strings.TrimSpace(m.Label), Form: outForm})
			}
			list = append(list, prov{ID: id, Label: p.Label, Configured: configured, Models: models})
		}
	} else {
		// Minimal fallback so UI still renders.
		list = append(list, prov{ID: "mock", Label: "Mock(联调)", Configured: true, Models: nil})
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
