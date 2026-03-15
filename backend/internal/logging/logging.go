package logging

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	once sync.Once
	lg   *slog.Logger
)

// InitFromEnv initializes the global logger once and sets slog.Default().
// Env:
// - LOG_LEVEL: debug|info|warn|error (default info)
// - LOG_FORMAT: json|text (default json)
func InitFromEnv() *slog.Logger {
	once.Do(func() {
		level := parseLevel(os.Getenv("LOG_LEVEL"))
		format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
		if format == "" {
			format = "json"
		}

		var h slog.Handler
		opts := &slog.HandlerOptions{Level: level}
		switch format {
		case "text":
			h = slog.NewTextHandler(os.Stdout, opts)
		default:
			h = slog.NewJSONHandler(os.Stdout, opts)
		}

		lg = slog.New(h)
		slog.SetDefault(lg)
	})
	return lg
}

func L() *slog.Logger {
	if lg == nil {
		return InitFromEnv()
	}
	return lg
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
