package httpapi

import (
	"net/http"
	"strings"
)

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := h.effectiveCfg()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"provider": strings.ToLower(strings.TrimSpace(cfg.Provider)),
	})
}

