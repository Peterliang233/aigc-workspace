package httpapi

import (
	"net/http"
)

func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.settingsGet(w, r)
		return
	case http.MethodPut:
		h.settingsPut(w, r)
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
