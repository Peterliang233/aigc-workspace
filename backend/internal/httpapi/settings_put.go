package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"aigc-backend/internal/settings"
)

type settingsPutReq struct {
	ImageProviders map[string]settings.ProviderSettings `json:"image_providers"`
}

func (h *Handler) settingsPut(w http.ResponseWriter, r *http.Request) {
	var req settingsPutReq
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.ImageProviders == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "image_providers is required"})
		return
	}

	known := map[string]bool{
		"openai_compatible": true,
		"siliconflow":       true,
		"wuyinkeji":         true,
	}

	_, err := h.st.Update(func(s *settings.Settings) error {
		if s.ImageProviders == nil {
			s.ImageProviders = map[string]settings.ProviderSettings{}
		}
		for id, patch := range req.ImageProviders {
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			if !known[id] {
				return fmt.Errorf("unknown provider: %s", id)
			}

			cur := s.ImageProviders[id]

			// base_url/api_key/default_model are env-managed (not editable via UI).
			if patch.BaseURL != nil || patch.APIKey != nil || patch.DefaultModel != nil {
				return fmt.Errorf("provider %s: base_url/api_key/default_model are env-managed", id)
			}
			if patch.Models != nil {
				var ms []string
				for _, m := range *patch.Models {
					m = strings.TrimSpace(m)
					if m != "" {
						ms = append(ms, m)
					}
				}
				cur.Models = &ms
			}

			s.ImageProviders[id] = cur
		}
		return nil
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "env-managed") {
			msg = "Base URL / API Key / 默认模型 需要通过部署环境配置"
		}
		slog.Default().Warn("settings_put_failed", "err", msg)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": msg})
		return
	}

	if err := h.reloadFromStore(); err != nil {
		slog.Default().Error("settings_reload_failed", "err", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	slog.Default().Info("settings_put_ok", "providers", len(req.ImageProviders))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

