package httpapi

import "strings"

func (h *Handler) defaultAudioProviderID() string {
	if h.models != nil {
		if id := strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("audio"))); id != "" {
			return id
		}
	}
	if pc, ok := h.cfg.ImageProviders["bltcy"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		return "bltcy"
	}
	return ""
}

func (h *Handler) getAudioProvider(id string) (audioProvider, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		id = h.defaultAudioProviderID()
	}
	if id == "" {
		return nil, false
	}
	h.provMu.Lock()
	defer h.provMu.Unlock()
	h.rebuildProvidersLocked()
	p, ok := h.audioProviders[id]
	return p, ok
}
