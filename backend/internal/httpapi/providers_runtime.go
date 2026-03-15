package httpapi

import (
	"log/slog"
	"strings"

	"aigc-backend/internal/config"
	"aigc-backend/internal/runtimecfg"
)

func (h *Handler) effectiveCfg() config.Config {
	h.cfgMu.RLock()
	defer h.cfgMu.RUnlock()
	return h.cfg
}

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

func (h *Handler) reloadFromStore() error {
	s, err := h.st.Get()
	if err != nil {
		return err
	}
	effective := runtimecfg.Merge(h.baseCfg, s)
	h.cfgMu.Lock()
	h.cfg = effective
	h.cfgMu.Unlock()

	h.provMu.Lock()
	h.imageProviders = map[string]imageProvider{}
	h.provKeys = map[string]string{}
	h.rebuildProvidersLocked()
	h.provMu.Unlock()

	slog.Default().Info("settings_reloaded")
	return nil
}

