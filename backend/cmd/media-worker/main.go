package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"aigc-backend/internal/mediaworker"
)

func main() {
	port := strings.TrimSpace(os.Getenv("MEDIA_WORKER_PORT"))
	if port == "" {
		port = "8090"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mediaworker.NewHandler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Minute,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	slog.Default().Info("media_worker_start", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Default().Error("media_worker_error", "err", err.Error())
		os.Exit(1)
	}
}
