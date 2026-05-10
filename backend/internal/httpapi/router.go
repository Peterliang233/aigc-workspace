package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/config"
	"aigc-backend/internal/modelcfg"
	"aigc-backend/internal/store"
	"aigc-backend/internal/storyvideo"

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
		storyMedia:     storyvideo.NewMediaClient(cfg.MediaWorkerURL),
		staticRoot:     staticRoot,
		imageProviders: map[string]imageProvider{},
		provKeys:       map[string]string{},
		videoProviders: map[string]videoProvider{},
		videoProvKeys:  map[string]string{},
		audioProviders: map[string]audioProvider{},
		audioProvKeys:  map[string]string{},
	}
	if strings.TrimSpace(cfg.MySQLDSN) != "" {
		if sv, err := storyvideo.NewStore(cfg.MySQLDSN); err != nil {
			slog.Default().Warn("story_video_store_init_failed", "err", err.Error())
		} else {
			h.storyVideos = sv
		}
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
	r.GET("/api/meta/texts", gin.WrapF(h.metaTexts))

	r.POST("/api/images/generate", gin.WrapF(h.imagesGenerate))
	r.POST("/api/audios/generate", gin.WrapF(h.audiosGenerate))
	r.POST("/api/texts/generate", gin.WrapF(h.textsGenerate))

	r.POST("/api/videos/jobs", gin.WrapF(h.videosJobs))
	r.GET("/api/videos/jobs/*id", gin.WrapF(h.videosJobsID))
	r.POST("/api/story-videos/projects/draft", gin.WrapF(h.storyVideoDraft))
	r.PUT("/api/story-videos/projects/:projectID/draft", gin.WrapF(h.storyVideoDraftID))
	r.POST("/api/story-videos/projects/:projectID/confirm", gin.WrapF(h.storyVideoConfirm))
	r.GET("/api/story-videos/projects", gin.WrapF(h.storyVideoProjects))
	r.GET("/api/story-videos/projects/:projectID", gin.WrapF(h.storyVideoProjectID))
	r.GET("/api/story-videos/projects/:projectID/events", gin.WrapF(h.storyVideoEvents))
	r.POST("/api/story-videos/projects/:projectID/compose", gin.WrapF(h.storyVideoCompose))
	r.POST("/api/story-videos/projects/:projectID/regenerate-audio", gin.WrapF(h.storyVideoRegenerateAudio))
	r.POST("/api/story-videos/projects/:projectID/shots/:shotID/regenerate-image", gin.WrapF(h.storyVideoRegenerateShotImage))

	r.GET("/api/history", gin.WrapF(h.historyList))
	r.GET("/api/history/*id", gin.WrapF(h.historyGet))
	r.DELETE("/api/history/*id", gin.WrapF(h.historyDelete))
	r.GET("/api/assets/*id", gin.WrapF(h.assetsGet))

	return r
}
