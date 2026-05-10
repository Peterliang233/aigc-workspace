package httpapi

import (
	"net/http"
	"sort"
	"strings"
)

func (h *Handler) metaTexts(w http.ResponseWriter, r *http.Request) {
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
		Placeholder string        `json:"placeholder,omitempty"`
		Required    bool          `json:"required,omitempty"`
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
		for _, provider := range h.models.Providers {
			if provider.Text == nil {
				continue
			}
			id := strings.ToLower(strings.TrimSpace(provider.ID))
			pc := h.cfg.ImageProviders[id]
			configured := strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != ""
			var models []model
			for _, item := range provider.Text.Models {
				out := model{ID: strings.TrimSpace(item.ID), Label: strings.TrimSpace(item.Label)}
				if item.Form != nil {
					out.Form = &form{}
					for _, f := range item.Form.Fields {
						row := field{Key: strings.TrimSpace(f.Key), Label: strings.TrimSpace(f.Label), Type: strings.TrimSpace(f.Type), Required: f.Required, Placeholder: strings.TrimSpace(f.Placeholder), Default: f.Default, Rows: f.Rows}
						for _, opt := range f.Options {
							row.Options = append(row.Options, fieldOption{Label: strings.TrimSpace(opt.Label), Value: strings.TrimSpace(opt.Value)})
						}
						if row.Key != "" && row.Type != "" {
							out.Form.Fields = append(out.Form.Fields, row)
						}
					}
				}
				if out.ID != "" {
					models = append(models, out)
				}
			}
			list = append(list, prov{ID: id, Label: provider.Label, Configured: configured, Models: models})
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	writeJSON(w, http.StatusOK, map[string]any{"default_provider": h.defaultTextProviderID(), "providers": list})
}
