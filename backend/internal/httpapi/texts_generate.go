package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/textgen"
	"aigc-backend/internal/types"
)

func (h *Handler) textsGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req types.TextGenerateRequest
	if err := decodeJSONWithLimit(w, r, &req, 2<<20); err != nil {
		return
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.SystemPrompt = strings.TrimSpace(req.SystemPrompt)
	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "prompt is required"})
		return
	}
	client, providerID, ok := h.textGeneratorFor(req.Provider)
	if !ok || client == nil || !client.Enabled() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "文本能力暂未启用或未配置"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	var resp textgen.Response
	err := retryDownstreamCall(ctx, "text_generate", func(callCtx context.Context) error {
		var callErr error
		resp, callErr = client.Generate(callCtx, textgen.Request{
			Model: req.Model, Prompt: req.Prompt, SystemPrompt: req.SystemPrompt,
			Temperature: req.Temperature, MaxTokens: req.MaxTokens,
		})
		return callErr
	})
	if err != nil {
		slog.Default().Warn("texts_generate_failed", "provider", providerID, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, types.TextGenerateResponse{Text: resp.Text, Provider: resp.Provider, Model: resp.Model})
}
