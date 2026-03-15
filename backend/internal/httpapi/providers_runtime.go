package httpapi

import (
	"strings"
)

func (h *Handler) getImageProvider(id string) (imageProvider, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return nil, false
	}

	// Ensure cache is up-to-date with current effective config.
	h.provMu.Lock()
	defer h.provMu.Unlock()
	h.rebuildProvidersLocked()
	p, ok := h.imageProviders[id]
	return p, ok
}
