package httpapi

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func (h *Handler) assetsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() || h.assets.MinIO == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	id = strings.Trim(id, "/")
	if id == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	a, err := h.assets.Store.Get(ctx, uid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	size := a.Bytes
	rangeHdr := strings.TrimSpace(r.Header.Get("Range"))
	start, end, partial := parseByteRange(rangeHdr, size)

	opts := minio.GetObjectOptions{}
	if partial {
		_ = opts.SetRange(start, end)
	}
	obj, err := h.assets.MinIO.Client.GetObject(ctx, h.assets.MinIO.Bucket, a.ObjectKey, opts)
	if err != nil {
		slog.Default().Warn("asset_get_failed", "id", uid, "err", err.Error())
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", a.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Accept-Ranges", "bytes")

	download := strings.TrimSpace(r.URL.Query().Get("download"))
	if download == "1" || strings.EqualFold(download, "true") {
		ext := ""
		if exts, _ := mime.ExtensionsByType(a.ContentType); len(exts) > 0 {
			ext = exts[0]
		}
		name := fmt.Sprintf("aigc_%s_%d%s", a.Capability, a.ID, ext)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	}

	if partial && size > 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		if size > 0 {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		}
		w.WriteHeader(http.StatusOK)
	}

	if _, err := io.Copy(w, obj); err != nil {
		slog.Default().Warn("asset_stream_failed", "id", uid, "err", err.Error())
	}
}

func parseByteRange(hdr string, size int64) (start, end int64, ok bool) {
	// Only support a single range, enough for HTML5 video playback.
	if size <= 0 {
		return 0, 0, false
	}
	hdr = strings.TrimSpace(hdr)
	if hdr == "" || !strings.HasPrefix(hdr, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(hdr, "bytes=")
	// Ignore multiple ranges.
	if strings.Contains(spec, ",") {
		return 0, 0, false
	}
	a, b, found := strings.Cut(spec, "-")
	if !found {
		return 0, 0, false
	}
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	// Suffix form: bytes=-N
	if a == "" {
		n, err := strconv.ParseInt(b, 10, 64)
		if err != nil || n <= 0 {
			return 0, 0, false
		}
		if n > size {
			n = size
		}
		return size - n, size - 1, true
	}

	st, err := strconv.ParseInt(a, 10, 64)
	if err != nil || st < 0 || st >= size {
		return 0, 0, false
	}
	en := size - 1
	if b != "" {
		v, err := strconv.ParseInt(b, 10, 64)
		if err != nil || v < st {
			return 0, 0, false
		}
		if v < size {
			en = v
		}
	}
	return st, en, true
}
