package httpapi

import (
	"net/http"
	"sort"
	"strings"
)

func (h *Handler) metaAudios(w http.ResponseWriter, r *http.Request) {
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
		Fields []field `json:"fields,omitempty"`
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
	if h.models != nil {
		for _, p := range h.models.Providers {
			if p.Audio == nil {
				continue
			}
			id := strings.ToLower(strings.TrimSpace(p.ID))
			if id == "" {
				continue
			}
			pc := h.cfg.ImageProviders[id]
			configured := strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != ""
			var models []model
			for _, m := range p.Audio.Models {
				mid := strings.TrimSpace(m.ID)
				if mid == "" {
					continue
				}
				var outForm *form
				if m.Form != nil {
					outForm = &form{}
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
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": h.defaultAudioProviderID(),
		"providers":        list,
	})
}
