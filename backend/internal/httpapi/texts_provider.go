package httpapi

import (
	"strings"

	"aigc-backend/internal/textgen"
)

func (h *Handler) defaultTextProviderID() string {
	if h.models != nil && strings.TrimSpace(h.models.DefaultProvider("text")) != "" {
		return h.models.DefaultProvider("text")
	}
	return "bltcy"
}

func (h *Handler) textGeneratorFor(providerID string) (*textgen.Client, string, bool) {
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	if providerID == "" {
		providerID = h.defaultTextProviderID()
	}
	pc, ok := h.cfg.ImageProviders[providerID]
	if !ok || strings.TrimSpace(pc.BaseURL) == "" || strings.TrimSpace(pc.APIKey) == "" {
		return nil, providerID, false
	}
	model := ""
	if h.models != nil {
		model = h.models.DefaultModel(providerID, "text")
	}
	return textgen.New(providerID, pc.BaseURL, pc.APIKey, model), providerID, true
}
