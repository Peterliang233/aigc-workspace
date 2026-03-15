package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) historyList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	capability := strings.TrimSpace(r.URL.Query().Get("capability"))
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))

	items, err := h.assets.Store.List(ctx, capability, limit, offset)
	if err != nil {
		slog.Default().Warn("history_list_failed", "err", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	type row struct {
		ID          uint64 `json:"id"`
		Capability  string `json:"capability"`
		Provider    string `json:"provider"`
		Model       string `json:"model,omitempty"`
		Status      string `json:"status"`
		Error       string `json:"error,omitempty"`
		Prompt      string `json:"prompt_preview,omitempty"`
		ContentType string `json:"content_type"`
		Bytes       int64  `json:"bytes"`
		URL         string `json:"url"`
		CreatedAt   string `json:"created_at"`
	}

	out := make([]row, 0, len(items))
	for _, a := range items {
		e := ""
		if a.Error != nil {
			e = *a.Error
		}
		out = append(out, row{
			ID:          a.ID,
			Capability:  a.Capability,
			Provider:    a.Provider,
			Model:       a.Model,
			Status:      a.Status,
			Error:       e,
			Prompt:      a.PromptPreview,
			ContentType: a.ContentType,
			Bytes:       a.Bytes,
			URL:         fmt.Sprintf("/api/assets/%d", a.ID),
			CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) historyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/history/")
	id = strings.Trim(id, "/")
	if id == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	a, err := h.assets.Store.Get(ctx, uid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}
	e := ""
	if a.Error != nil {
		e = *a.Error
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":             a.ID,
		"capability":     a.Capability,
		"provider":       a.Provider,
		"model":          a.Model,
		"status":         a.Status,
		"error":          e,
		"prompt_preview": a.PromptPreview,
		"content_type":   a.ContentType,
		"bytes":          a.Bytes,
		"url":            fmt.Sprintf("/api/assets/%d", a.ID),
		"created_at":     a.CreatedAt.Format(time.RFC3339),
	})
}
