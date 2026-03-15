package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"aigc-backend/internal/config"
	"aigc-backend/internal/httpapi"
)

func main() {
	// Convenience for local dev: load backend/.env if present.
	// Environment variables set by the shell take precedence.
	config.LoadDotEnv(".env")
	cfg := config.LoadFromEnv()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewHandler(cfg),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("backend listening on http://localhost:%s (provider=%s)\n", cfg.Port, cfg.Provider)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
