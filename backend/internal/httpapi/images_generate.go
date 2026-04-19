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
	"aigc-backend/internal/logging"
	"aigc-backend/internal/types"
)

func (h *Handler) imagesGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	var req types.ImageGenerateRequest
	if err := decodeJSONWithLimit(w, r, &req, 20<<20); err != nil {
		return
	}
	// Current UI + some providers only support a single output.
	req.N = 1

	providerID := strings.ToLower(strings.TrimSpace(req.Provider))
	if providerID == "" {
		if h.models != nil {
			providerID = strings.ToLower(strings.TrimSpace(h.models.DefaultProvider("image")))
		}
	}
	if providerID == "" {
		providerID = "mock"
	}
	prov, ok := h.getImageProvider(providerID)
	if !ok || prov == nil {
		slog.Default().Warn("images_generate_unknown_provider", "provider", providerID)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unknown provider: " + providerID})
		return
	}

	// Apply per-provider default model from models.json when request doesn't specify one.
	if strings.TrimSpace(req.Model) == "" && h.models != nil {
		req.Model = h.models.DefaultModel(providerID, "image")
	}
	h.applyImageModelDefaults(providerID, strings.TrimSpace(req.Model), &req)
	if miss := h.missingImageRequiredFields(providerID, strings.TrimSpace(req.Model), req); len(miss) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing required fields: " + strings.Join(miss, ", ")})
		return
	}

	slog.Default().Info("images_generate",
		"provider", providerID,
		"model", strings.TrimSpace(req.Model),
		"size", strings.TrimSpace(req.Size),
		"n", req.N,
	)
	var resp types.ImageGenerateResponse
	err := retryDownstreamCall(ctx, "image_generate", func(callCtx context.Context) error {
		var genErr error
		resp, genErr = prov.GenerateImage(callCtx, req)
		return genErr
	})
	if err != nil {
		slog.Default().Warn("images_generate_failed", "provider", providerID, "err", err.Error())
		msg := err.Error()
		if strings.Contains(msg, "env-managed") {
			msg = "Base URL / API Key / 默认模型 需要通过部署环境配置"
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": msg})
		return
	}
	if strings.TrimSpace(resp.Provider) == "" {
		resp.Provider = providerID
	}

	// Persist the first image into MinIO and rewrite response URL to this backend.
	if h.assets != nil && h.assets.Enabled() && len(resp.ImageURLs) > 0 {
		src := strings.TrimSpace(resp.ImageURLs[0])
		if src != "" {
			var a *assets.Asset
			var err error

			// Some providers (mock/siliconflow) return local URLs under /static/generated.
			// Upload those files into MinIO so history works consistently.
			if strings.HasPrefix(src, "/static/generated/") {
				p := filepath.Join(h.staticRoot, "generated", filepath.Base(src))
				a, err = h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
					Capability: "image",
					Provider:   providerID,
					Model:      strings.TrimSpace(resp.Model),
					Prompt:     req.Prompt,
					Params: map[string]any{
						"size": strings.TrimSpace(req.Size),
					},
					FilePath: p,
				})
				if err == nil && a != nil {
					// Best-effort cleanup: this file is only a staging artifact once MinIO is enabled.
					if rmErr := os.Remove(p); rmErr != nil {
						slog.Default().Warn("images_cleanup_local_failed", "path", p, "err", rmErr.Error())
					}
				}
			} else {
				a, err = h.assets.StoreRemote(ctx, assets.StoreRemoteInput{
					Capability: "image",
					Provider:   providerID,
					Model:      strings.TrimSpace(resp.Model),
					Prompt:     req.Prompt,
					Params: map[string]any{
						"size": strings.TrimSpace(req.Size),
					},
					SourceURL: src,
				})
			}

			if err != nil {
				slog.Default().Warn("images_store_asset_failed", "provider", providerID, "src", logging.RedactURL(src), "err", err.Error())
			} else if a != nil {
				resp.ImageURLs = []string{fmt.Sprintf("/api/assets/%d", a.ID)}
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
