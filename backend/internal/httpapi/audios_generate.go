package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/types"
)

func (h *Handler) audiosGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	var req types.AudioGenerateRequest
	if err := decodeJSONWithLimit(w, r, &req, 2<<20); err != nil {
		return
	}
	providerID := strings.ToLower(strings.TrimSpace(req.Provider))
	if providerID == "" && h.models != nil {
		providerID = strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("audio")))
	}
	prov, ok := h.getAudioProvider(providerID)
	if !ok || prov == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "音频能力暂未启用或未配置"})
		return
	}
	if strings.TrimSpace(req.Model) == "" && h.models != nil {
		req.Model = h.models.DefaultModel(providerID, "audio")
	}
	h.applyAudioModelDefaults(providerID, strings.TrimSpace(req.Model), &req)
	if miss := h.missingAudioRequiredFields(providerID, strings.TrimSpace(req.Model), req); len(miss) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing required fields: " + strings.Join(miss, ", ")})
		return
	}

	resp, err := prov.GenerateAudio(ctx, req)
	if err != nil {
		slog.Default().Warn("audios_generate_failed", "provider", providerID, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if strings.TrimSpace(resp.Provider) == "" {
		resp.Provider = providerID
	}
	if h.assets != nil && h.assets.Enabled() && strings.HasPrefix(resp.AudioURL, "/static/generated/") {
		p := filepath.Join(h.staticRoot, "generated", filepath.Base(resp.AudioURL))
		a, err := h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
			Capability: "audio",
			Provider:   providerID,
			Model:      strings.TrimSpace(resp.Model),
			Prompt:     req.Input,
			Params: map[string]any{
				"voice":           strings.TrimSpace(req.Voice),
				"response_format": strings.TrimSpace(req.ResponseFormat),
				"speed":           req.Speed,
			},
			FilePath:    p,
			ContentType: resp.ContentType,
		})
		if err != nil {
			slog.Default().Warn("audios_store_asset_failed", "provider", providerID, "err", err.Error())
		} else if a != nil {
			resp.AudioURL = fmt.Sprintf("/api/assets/%d", a.ID)
			_ = os.Remove(p)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
