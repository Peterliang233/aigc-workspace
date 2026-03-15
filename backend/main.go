package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/blobstore"
	"aigc-backend/internal/config"
	"aigc-backend/internal/httpapi"
	"aigc-backend/internal/logging"
	"aigc-backend/internal/modelcfg"
)

func main() {
	// Convenience for local dev: load the first .env found when walking upwards from CWD.
	// Environment variables set by the shell take precedence.
	config.LoadDotEnvUpwards(8)
	cfg := config.LoadFromEnv()

	logging.InitFromEnv()

	if strings.TrimSpace(cfg.MySQLDSN) == "" {
		slog.Default().Error("mysql_missing_dsn")
		os.Exit(1)
	}

	modelsPath := strings.TrimSpace(os.Getenv("MODELS_CONFIG_PATH"))
	if modelsPath == "" {
		modelsPath = "models.json"
	}
	models, err := modelcfg.Load(modelsPath)
	if err != nil {
		// Give a more actionable path hint for common layouts.
		slog.Default().Error("models_config_load_failed",
			"path", modelsPath,
			"cwd_hint", filepath.Base(mustGetwd()),
			"err", err.Error(),
		)
		os.Exit(1)
	}

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
		Handler:           httpapi.NewHandler(cfg, models, assetSvc),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	slog.Default().Info("server_start",
		"addr", srv.Addr,
		"mysql", cfg.MySQLDSN != "",
		"minio", minioStore != nil,
		"models_version", models.Version,
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Default().Error("server_error", "err", err.Error())
		os.Exit(1)
	}
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}
