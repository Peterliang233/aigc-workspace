package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func slogHTTPMiddleware(allowedOrigins []string) gin.HandlerFunc {
	corsAllowed := map[string]bool{}
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			corsAllowed[o] = true
		}
	}

	return func(c *gin.Context) {
		start := time.Now()

		rid := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if rid == "" {
			rid = newRequestID()
		}
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Set("rid", rid)

		// CORS (kept compatible with current front-end)
		origin := c.GetHeader("Origin")
		if origin != "" && corsAllowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}

		defer func() {
			if rec := recover(); rec != nil {
				slog.Default().Error("panic",
					"rid", rid,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"recover", rec,
				)
				// Keep it plain-text and generic.
				c.Writer.WriteHeader(500)
			}

			dur := time.Since(start)
			slog.Default().Info("http",
				"rid", rid,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"bytes", c.Writer.Size(),
				"dur_ms", dur.Milliseconds(),
				"ip", c.ClientIP(),
				"ua", c.Request.UserAgent(),
			)
		}()

		c.Next()
	}
}

