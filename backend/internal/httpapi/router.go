package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/config"
	"aigc-backend/internal/modelcfg"
	"aigc-backend/internal/store"

	"github.com/gin-gonic/gin"
)

func NewHandler(cfg config.Config, models *modelcfg.Config, assetsSvc *assets.Service) http.Handler {
	staticRoot := filepath.Join("var")
	_ = os.MkdirAll(filepath.Join(staticRoot, "generated"), 0o755)

	slog.Default().Info("handler_init",
		"static_root", staticRoot,
		"models_version", func() int {
			if models != nil {
				return models.Version
			}
			return 0
		}(),
		"allowed_origins", len(cfg.AllowedOrigins),
	)

	h := &Handler{
		cfg:            cfg,
		models:         models,
		assets:         assetsSvc,
		jobs:           store.NewJobStore(),
		staticRoot:     staticRoot,
		imageProviders: map[string]imageProvider{},
		provKeys:       map[string]string{},
		videoProviders: map[string]videoProvider{},
		videoProvKeys:  map[string]string{},
		audioProviders: map[string]audioProvider{},
		audioProvKeys:  map[string]string{},
	}
	h.provMu.Lock()
	h.rebuildProvidersLocked()
	h.provMu.Unlock()

	r := gin.New()
	r.Use(slogHTTPMiddleware(cfg.AllowedOrigins))

	r.GET("/healthz", gin.WrapF(h.healthz))

	// Only expose generated media. Do not expose the whole var/ directory, because it may contain secrets/config.
	r.StaticFS("/static/generated", http.Dir(filepath.Join(staticRoot, "generated")))

	r.GET("/api/meta/images", gin.WrapF(h.metaImages))
	r.GET("/api/meta/videos", gin.WrapF(h.metaVideos))
	r.GET("/api/meta/audios", gin.WrapF(h.metaAudios))

	r.POST("/api/images/generate", gin.WrapF(h.imagesGenerate))
	r.POST("/api/audios/generate", gin.WrapF(h.audiosGenerate))

	r.POST("/api/videos/jobs", gin.WrapF(h.videosJobs))
	r.GET("/api/videos/jobs/*id", gin.WrapF(h.videosJobsID))

	r.GET("/api/history", gin.WrapF(h.historyList))
	r.GET("/api/history/*id", gin.WrapF(h.historyGet))
	r.DELETE("/api/history/*id", gin.WrapF(h.historyDelete))
	r.GET("/api/assets/*id", gin.WrapF(h.assetsGet))

	return r
}
