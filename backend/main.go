package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/blobstore"
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

	assetStore, err := assets.NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		slog.Default().Error("assets_store_mysql_init_failed", "err", err.Error())
		os.Exit(1)
	}
	defer assetStore.Close()

	var minioStore *blobstore.MinIO
	if strings.TrimSpace(cfg.MinIOEndpoint) != "" {
		minioStore, err = blobstore.NewMinIO(blobstore.MinIOConfig{
			Endpoint:  cfg.MinIOEndpoint,
			AccessKey: cfg.MinIOAccessKey,
			SecretKey: cfg.MinIOSecretKey,
			Bucket:    cfg.MinIOBucket,
			UseSSL:    cfg.MinIOUseSSL,
		})
		if err != nil {
			slog.Default().Error("minio_init_failed", "err", err.Error())
			os.Exit(1)
		}
	} else {
		slog.Default().Warn("minio_disabled", "reason", "MINIO_ENDPOINT is empty")
	}

	assetSvc := &assets.Service{
		Store: assetStore,
		MinIO: minioStore,
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewHandler(cfg, st, assetSvc),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	slog.Default().Info("server_start",
		"addr", srv.Addr,
		"provider", cfg.Provider,
		"mysql", cfg.MySQLDSN != "",
		"minio", minioStore != nil,
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Default().Error("server_error", "err", err.Error())
		os.Exit(1)
	}
}
