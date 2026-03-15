package httpapi

import "strings"

func (h *Handler) modelRequiresInitImage(providerID, model string) bool {
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	model = strings.TrimSpace(model)
	if providerID != "" && model != "" && h.models != nil {
		if ms := h.models.Model(providerID, "video", model); ms != nil && ms.Form != nil {
			return ms.Form.RequiresImage
		}
	}

	// Fallback heuristic for unknown models (keeps behavior for SiliconFlow I2V families).
	u := strings.ToUpper(model)
	return strings.Contains(u, "I2V") || strings.Contains(u, "IMG2VID") || strings.Contains(u, "IMAGE2VIDEO")
}
