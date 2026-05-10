package httpapi

import (
	"net/http"
	"strconv"

	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.storyVideos.ListProjects(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	items := make([]types.StoryVideoProject, 0, len(rows))
	for i := range rows {
		items = append(items, h.storyVideoResponse(&rows[i], nil))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) storyVideoProjectID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	project, shots, err := h.storyVideos.GetProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, h.storyVideoResponse(project, shots))
}

func (h *Handler) storyVideoEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.storyVideoReady(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
		return
	}
	projectID := storyVideoProjectID(r.URL.Path)
	events, err := h.storyVideos.ListEvents(r.Context(), projectID, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	items := make([]types.StoryVideoEvent, 0, len(events))
	for _, event := range events {
		items = append(items, storyVideoEventResponse(event))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
