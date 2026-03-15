package httpapi

import (
	"net/http"
)

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mv := 0
	if h.models != nil {
		mv = h.models.Version
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "models_version": mv})
}
