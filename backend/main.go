package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"aigc-backend/internal/config"
	"aigc-backend/internal/httpapi"
	"aigc-backend/internal/logging"
	"aigc-backend/internal/settings"
)

func main() {
	// Convenience for local dev: load the first .env found when walking upwards from CWD.
	// Environment variables set by the shell take precedence.
	config.LoadDotEnvUpwards(8)
	cfg := config.LoadFromEnv()

	logging.InitFromEnv()

	var st settings.Store
	if strings.TrimSpace(cfg.MySQLDSN) == "" {
		slog.Default().Error("settings_store_mysql_missing_dsn")
		os.Exit(1)
	}
	mysqlStore, err := settings.NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		slog.Default().Error("settings_store_mysql_init_failed", "err", err.Error())
		os.Exit(1)
	}
	st = mysqlStore
	defer st.Close()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewHandler(cfg, st),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	slog.Default().Info("server_start",
		"addr", srv.Addr,
		"provider", cfg.Provider,
		"mysql", cfg.MySQLDSN != "",
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Default().Error("server_error", "err", err.Error())
		os.Exit(1)
	}
}
