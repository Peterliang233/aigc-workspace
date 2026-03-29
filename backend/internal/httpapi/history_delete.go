package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func (h *Handler) historyDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() || h.assets.MinIO == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "history storage not enabled"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/history/")
	id = strings.Trim(id, "/")
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil || uid == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	a, err := h.assets.Store.Get(ctx, uid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}

	objKey := strings.TrimSpace(a.ObjectKey)
	if objKey != "" {
		if err := h.assets.MinIO.Client.RemoveObject(ctx, h.assets.MinIO.Bucket, objKey, minio.RemoveObjectOptions{}); err != nil {
			slog.Default().Warn("history_delete_object_failed", "id", uid, "object_key", objKey, "err", err.Error())
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "删除对象失败: " + err.Error()})
			return
		}
	}
	if err := h.assets.Store.Delete(ctx, uid); err != nil {
		slog.Default().Warn("history_delete_db_failed", "id", uid, "err", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除记录失败: " + err.Error()})
		return
	}

	slog.Default().Info("history_deleted", "id", uid, "object_key", objKey)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": uid})
}
