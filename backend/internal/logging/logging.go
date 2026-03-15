package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	once sync.Once
	lg   *slog.Logger
	logFile *os.File
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

		out := io.Writer(os.Stdout)
		if f, err := openLogFileFromEnv(); err == nil && f != nil {
			logFile = f
			out = io.MultiWriter(os.Stdout, f)
		} else if err != nil {
			// Logger not initialized yet; write to stderr best-effort.
			_, _ = fmt.Fprintf(os.Stderr, "logging: failed to open log file: %v\n", err)
		}

		var h slog.Handler
		opts := &slog.HandlerOptions{Level: level}
		switch format {
		case "text":
			h = slog.NewTextHandler(out, opts)
		default:
			h = slog.NewJSONHandler(out, opts)
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

func openLogFileFromEnv() (*os.File, error) {
	// Default: write to <repoRoot>/log/backend-YYYYMMDD.log.
	// If repo root cannot be detected, use ./log/...
	root := findRepoRootUpwards(10)
	if root == "" {
		root, _ = os.Getwd()
	}

	path := strings.TrimSpace(os.Getenv("LOG_FILE"))
	if path == "" {
		name := "backend-" + time.Now().Format("20060102") + ".log"
		path = filepath.Join(root, "log", name)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
}

func findRepoRootUpwards(maxHops int) string {
	if maxHops <= 0 {
		maxHops = 8
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < maxHops; i++ {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
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
