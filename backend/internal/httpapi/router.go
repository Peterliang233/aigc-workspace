package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/config"
	"aigc-backend/internal/runtimecfg"
	"aigc-backend/internal/settings"
	"aigc-backend/internal/store"

	"github.com/gin-gonic/gin"
)

func NewHandler(cfg config.Config, st settings.Store, assetsSvc *assets.Service) http.Handler {
	staticRoot := filepath.Join("var")
	_ = os.MkdirAll(filepath.Join(staticRoot, "generated"), 0o755)

	s, _ := st.Get()
	effective := runtimecfg.Merge(cfg, s)

	slog.Default().Info("handler_init",
		"static_root", staticRoot,
		"default_provider", strings.ToLower(strings.TrimSpace(effective.Provider)),
		"allowed_origins", len(cfg.AllowedOrigins),
	)

	h := &Handler{
		baseCfg:        cfg,
		st:             st,
		assets:         assetsSvc,
		cfg:            effective,
		jobs:           store.NewJobStore(),
		staticRoot:     staticRoot,
		imageProviders: map[string]imageProvider{},
		provKeys:       map[string]string{},
		videoProviders: map[string]videoProvider{},
		videoProvKeys:  map[string]string{},
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

	r.GET("/api/settings", gin.WrapF(h.settings))
	r.PUT("/api/settings", gin.WrapF(h.settings))
	r.POST("/api/settings/image-providers/*path", gin.WrapF(h.settingsImageProviders))
	r.DELETE("/api/settings/image-providers/*path", gin.WrapF(h.settingsImageProviders))

	r.POST("/api/images/generate", gin.WrapF(h.imagesGenerate))

	r.POST("/api/videos/jobs", gin.WrapF(h.videosJobs))
	r.GET("/api/videos/jobs/*id", gin.WrapF(h.videosJobsID))

	r.GET("/api/history", gin.WrapF(h.historyList))
	r.GET("/api/history/*id", gin.WrapF(h.historyGet))
	r.GET("/api/assets/*id", gin.WrapF(h.assetsGet))

	return r
}
