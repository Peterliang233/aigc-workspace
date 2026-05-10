package httpapi

import (
	"context"
	"encoding/json"

	"aigc-backend/internal/storyvideo"
)

func (h *Handler) storyVideoAddEvent(ctx context.Context, projectID, stage, typ, message string, payload any) error {
	if h.storyVideos == nil {
		return nil
	}
	raw := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			raw = string(b)
		}
	}
	return h.storyVideos.AddEvent(ctx, &storyvideo.Event{
		ProjectID: projectID,
		Stage:     stage,
		Type:      typ,
		Message:   message,
		Payload:   raw,
	})
}
