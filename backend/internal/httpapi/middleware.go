package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type ctxKey string

const ctxKeyRequestID ctxKey = "request_id"

func requestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(ctxKeyRequestID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func remoteIP(r *http.Request) string {
	// Prefer X-Forwarded-For when behind nginx.
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func withMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	corsAllowed := map[string]bool{}
	for _, o := range allowedOrigins {
		corsAllowed[strings.TrimSpace(o)] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rid := r.Header.Get("X-Request-ID")
		if strings.TrimSpace(rid) == "" {
			rid = newRequestID()
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, rid)
		r = r.WithContext(ctx)

		// CORS (kept compatible with current front-end)
		origin := r.Header.Get("Origin")
		if origin != "" && corsAllowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		sw := &statusWriter{ResponseWriter: w}

		defer func() {
			if rec := recover(); rec != nil {
				slog.Default().Error("panic", "rid", rid, "path", r.URL.Path, "method", r.Method, "recover", rec)
				http.Error(sw, "internal server error", http.StatusInternalServerError)
			}

			dur := time.Since(start)
			slog.Default().Info("http",
				"rid", rid,
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"bytes", sw.bytes,
				"dur_ms", dur.Milliseconds(),
				"ip", remoteIP(r),
				"ua", r.UserAgent(),
			)
		}()

		next.ServeHTTP(sw, r)
	})
}
