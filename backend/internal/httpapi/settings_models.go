package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"aigc-backend/internal/providers/wuyinkeji"
	"aigc-backend/internal/settings"
)

func (h *Handler) settingsImageProviders(w http.ResponseWriter, r *http.Request) {
	// Routes:
	// - POST   /api/settings/image-providers/{provider}/models   { "model": "..." }
	// - DELETE /api/settings/image-providers/{provider}/models?model=...
	// - DELETE /api/settings/image-providers/{provider}
	path := strings.TrimPrefix(r.URL.Path, "/api/settings/image-providers/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	parts := strings.Split(path, "/")
	providerID := strings.ToLower(strings.TrimSpace(parts[0]))
	if providerID == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	known := map[string]bool{
		"openai_compatible": true,
		"siliconflow":       true,
		"wuyinkeji":         true,
	}
	if !known[providerID] {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unknown provider: " + providerID})
		return
	}

	if len(parts) == 1 {
		h.settingsProviderReset(w, r, providerID)
		return
	}
	if len(parts) == 2 && parts[1] == "models" {
		h.settingsModels(w, r, providerID)
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (h *Handler) settingsProviderReset(w http.ResponseWriter, r *http.Request, providerID string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := h.st.Update(func(s *settings.Settings) error {
		if s.ImageProviders == nil {
			return nil
		}
		delete(s.ImageProviders, providerID)
		return nil
	})
	if err != nil {
		slog.Default().Warn("settings_provider_reset_failed", "provider", providerID, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if err := h.reloadFromStore(); err != nil {
		slog.Default().Error("settings_reload_failed", "err", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	slog.Default().Info("settings_provider_reset", "provider", providerID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) settingsModels(w http.ResponseWriter, r *http.Request, providerID string) {
	switch r.Method {
	case http.MethodPost:
		var body struct {
			Model string `json:"model"`
		}
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		model := strings.TrimSpace(body.Model)
		if model == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "model is required"})
			return
		}

		baseModels := h.currentModels(providerID)
		_, err := h.st.Update(func(s *settings.Settings) error {
			if s.ImageProviders == nil {
				s.ImageProviders = map[string]settings.ProviderSettings{}
			}
			cur := s.ImageProviders[providerID]

			// If there is no explicit models override yet, start from current effective list,
			// so "add" doesn't accidentally wipe env-provided models.
			var list []string
			if cur.Models != nil {
				list = append(list, (*cur.Models)...)
			} else {
				list = append(list, baseModels...)
			}
			if !containsStr(list, model) {
				list = append(list, model)
			}
			cur.Models = &list
			s.ImageProviders[providerID] = cur
			return nil
		})
		if err != nil {
			slog.Default().Warn("settings_model_add_failed", "provider", providerID, "model", model, "err", err.Error())
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if err := h.reloadFromStore(); err != nil {
			slog.Default().Error("settings_reload_failed", "err", err.Error())
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		slog.Default().Info("settings_model_added", "provider", providerID, "model", model)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return

	case http.MethodDelete:
		model := strings.TrimSpace(r.URL.Query().Get("model"))
		if model == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "model query param is required"})
			return
		}

		baseModels := h.currentModels(providerID)
		_, err := h.st.Update(func(s *settings.Settings) error {
			if s.ImageProviders == nil {
				s.ImageProviders = map[string]settings.ProviderSettings{}
			}
			cur := s.ImageProviders[providerID]

			var list []string
			if cur.Models != nil {
				list = append(list, (*cur.Models)...)
			} else {
				list = append(list, baseModels...)
			}
			list = removeStr(list, model)
			cur.Models = &list
			s.ImageProviders[providerID] = cur
			return nil
		})
		if err != nil {
			slog.Default().Warn("settings_model_delete_failed", "provider", providerID, "model", model, "err", err.Error())
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if err := h.reloadFromStore(); err != nil {
			slog.Default().Error("settings_reload_failed", "err", err.Error())
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		slog.Default().Info("settings_model_deleted", "provider", providerID, "model", model)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) currentModels(providerID string) []string {
	cfg := h.effectiveCfg()
	pc, ok := cfg.ImageProviders[providerID]
	if !ok {
		return nil
	}
	models := pc.Models
	if providerID == "wuyinkeji" {
		if provImpl, ok := h.getImageProvider("wuyinkeji"); ok {
			if p, ok := provImpl.(*wuyinkeji.Provider); ok {
				models = p.Models()
			}
		}
	}
	return append([]string{}, models...)
}

