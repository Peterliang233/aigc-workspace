package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/config"
	"aigc-backend/internal/providers/mock"
	"aigc-backend/internal/providers/openai_compatible"
	"aigc-backend/internal/providers/siliconflow"
	"aigc-backend/internal/providers/wuyinkeji"
	"aigc-backend/internal/runtimecfg"
	"aigc-backend/internal/settings"
	"aigc-backend/internal/store"
	"aigc-backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

type Handler struct {
	baseCfg config.Config
	st      settings.Store
	assets  *assets.Service

	cfgMu sync.RWMutex
	cfg   config.Config

	provMu         sync.Mutex
	imageProviders map[string]imageProvider
	provKeys       map[string]string
	videoProv      interface {
		ProviderName() string
		StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error)
		GetVideoJob(ctx context.Context, jobID string) (string, string, string, error)
	}

	jobs       *store.JobStore
	staticRoot string
}

type imageProvider interface {
	ProviderName() string
	GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error)
}

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

	r.GET("/api/settings", gin.WrapF(h.settings))
	r.PUT("/api/settings", gin.WrapF(h.settings))

	// Keep the legacy path-parsing handler, but route through gin.
	r.POST("/api/settings/image-providers/*path", gin.WrapF(h.settingsImageProviders))
	r.DELETE("/api/settings/image-providers/*path", gin.WrapF(h.settingsImageProviders))

	r.POST("/api/images/generate", gin.WrapF(h.imagesGenerate))

	r.POST("/api/videos/jobs", gin.WrapF(h.videosJobs))
	r.GET("/api/videos/jobs/*id", gin.WrapF(h.videosJobsID)) // GET /api/videos/jobs/{id}

	// Generation history + assets (MinIO-backed).
	r.GET("/api/history", gin.WrapF(h.historyList))
	r.GET("/api/history/*id", gin.WrapF(h.historyGet))
	r.GET("/api/assets/*id", gin.WrapF(h.assetsGet))

	return r
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := h.effectiveCfg()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"provider": strings.ToLower(strings.TrimSpace(cfg.Provider)),
	})
}

func (h *Handler) metaImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type prov struct {
		ID         string   `json:"id"`
		Label      string   `json:"label"`
		Configured bool     `json:"configured"`
		Models     []string `json:"models"`
	}

	labels := map[string]string{
		"mock":              "Mock(联调)",
		"openai_compatible": "OpenAI Compatible",
		"siliconflow":       "SiliconFlow",
		"wuyinkeji":         "无印科技(速创API)",
	}

	var list []prov
	cfg := h.effectiveCfg()
	for id, pc := range cfg.ImageProviders {
		// Only expose providers we actually have implementations for.
		if _, ok := labels[id]; !ok {
			continue
		}

		models := pc.Models
		// wuyinkeji may have legacy endpoint mapping; prefer provider-exposed model list for UI.
		if id == "wuyinkeji" {
			if prov, ok := h.getImageProvider("wuyinkeji"); ok {
				if p, ok := prov.(*wuyinkeji.Provider); ok {
					models = p.Models()
				}
			}
		}

		configured := true
		if id != "mock" && strings.TrimSpace(pc.APIKey) == "" {
			configured = false
		}

		list = append(list, prov{
			ID:         id,
			Label:      labels[id],
			Configured: configured,
			Models:     models,
		})
	}
	// Ensure mock exists even if cfg.ImageProviders was nil/unset.
	if _, ok := labels["mock"]; ok {
		found := false
		for _, p := range list {
			if p.ID == "mock" {
				found = true
				break
			}
		}
		if !found {
			list = append(list, prov{ID: "mock", Label: labels["mock"], Configured: true})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": strings.ToLower(strings.TrimSpace(cfg.Provider)),
		"providers":        list,
	})

	slog.Default().Debug("meta_images",
		"default_provider", strings.ToLower(strings.TrimSpace(cfg.Provider)),
		"providers", len(list),
	)
}

func (h *Handler) imagesGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	var req types.ImageGenerateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	// Current UI + some providers only support a single output.
	// We still accept `n` in the request for compatibility, but force 1 server-side.
	req.N = 1

	cfg := h.effectiveCfg()
	providerID := strings.ToLower(strings.TrimSpace(req.Provider))
	if providerID == "" {
		providerID = strings.ToLower(strings.TrimSpace(cfg.Provider))
	}
	if providerID == "" {
		providerID = "mock"
	}
	prov, ok := h.getImageProvider(providerID)
	if !ok || prov == nil {
		slog.Default().Warn("images_generate_unknown_provider", "provider", providerID)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unknown provider: " + providerID})
		return
	}

	slog.Default().Info("images_generate",
		"provider", providerID,
		"model", strings.TrimSpace(req.Model),
		"size", strings.TrimSpace(req.Size),
		"n", req.N,
	)
	resp, err := prov.GenerateImage(ctx, req)
	if err != nil {
		slog.Default().Warn("images_generate_failed", "provider", providerID, "err", err.Error())
		msg := err.Error()
		if strings.Contains(msg, "env-managed") {
			msg = "Base URL / API Key / 默认模型 需要通过部署环境配置"
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": msg})
		return
	}
	// Ensure provider field present even if provider implementation left it blank.
	if strings.TrimSpace(resp.Provider) == "" {
		resp.Provider = providerID
	}

	// Persist the first image into MinIO and rewrite response URL to this backend.
	if h.assets != nil && h.assets.Enabled() && len(resp.ImageURLs) > 0 {
		src := strings.TrimSpace(resp.ImageURLs[0])
		if src != "" {
			a, err := h.assets.StoreRemote(ctx, assets.StoreRemoteInput{
				Capability: "image",
				Provider:   providerID,
				Model:      strings.TrimSpace(resp.Model),
				Prompt:     req.Prompt,
				Params: map[string]any{
					"size": strings.TrimSpace(req.Size),
				},
				SourceURL: src,
			})
			if err != nil {
				slog.Default().Warn("images_store_asset_failed", "provider", providerID, "err", err.Error())
			} else {
				resp.ImageURLs = []string{fmt.Sprintf("/api/assets/%d", a.ID)}
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.settingsGet(w, r)
		return
	case http.MethodPut:
		h.settingsPut(w, r)
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) settingsGet(w http.ResponseWriter, r *http.Request) {
	type prov struct {
		Label        string   `json:"label"`
		BaseURL      string   `json:"base_url,omitempty"`
		APIKeySet    bool     `json:"api_key_set"`
		DefaultModel string   `json:"default_model,omitempty"`
		Models       []string `json:"models,omitempty"`
	}

	labels := map[string]string{
		"openai_compatible": "OpenAI Compatible",
		"siliconflow":       "SiliconFlow",
		"wuyinkeji":         "无印科技(速创API)",
	}

	cfg := h.effectiveCfg()
	out := map[string]prov{}
	for id, label := range labels {
		pc := cfg.ImageProviders[id]
		models := pc.Models
		if id == "wuyinkeji" {
			if provImpl, ok := h.getImageProvider("wuyinkeji"); ok {
				if p, ok := provImpl.(*wuyinkeji.Provider); ok {
					models = p.Models()
				}
			}
		}

		out[id] = prov{
			Label:        label,
			BaseURL:      strings.TrimSpace(pc.BaseURL),
			APIKeySet:    strings.TrimSpace(pc.APIKey) != "",
			DefaultModel: strings.TrimSpace(pc.DefaultModel),
			Models:       models,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"image_providers": out,
	})
}

type settingsPutReq struct {
	ImageProviders map[string]settings.ProviderSettings `json:"image_providers"`
}

func (h *Handler) settingsPut(w http.ResponseWriter, r *http.Request) {
	var req settingsPutReq
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.ImageProviders == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "image_providers is required"})
		return
	}

	known := map[string]bool{
		"openai_compatible": true,
		"siliconflow":       true,
		"wuyinkeji":         true,
	}

	_, err := h.st.Update(func(s *settings.Settings) error {
		if s.ImageProviders == nil {
			s.ImageProviders = map[string]settings.ProviderSettings{}
		}
		for id, patch := range req.ImageProviders {
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			if !known[id] {
				return fmt.Errorf("unknown provider: %s", id)
			}

			cur := s.ImageProviders[id]

			// base_url/api_key/default_model are env-managed (not editable via UI).
			if patch.BaseURL != nil || patch.APIKey != nil || patch.DefaultModel != nil {
				return fmt.Errorf("provider %s: base_url/api_key/default_model are env-managed", id)
			}
			if patch.Models != nil {
				var ms []string
				for _, m := range *patch.Models {
					m = strings.TrimSpace(m)
					if m != "" {
						ms = append(ms, m)
					}
				}
				cur.Models = &ms
			}

			s.ImageProviders[id] = cur
		}
		return nil
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "env-managed") {
			msg = "Base URL / API Key / 默认模型 需要通过部署环境配置"
		}
		slog.Default().Warn("settings_put_failed", "err", msg)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": msg})
		return
	}

	if err := h.reloadFromStore(); err != nil {
		slog.Default().Error("settings_reload_failed", "err", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	slog.Default().Info("settings_put_ok", "providers", len(req.ImageProviders))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) settingsImageProviders(w http.ResponseWriter, r *http.Request) {
	// Routes:
	// - POST   /api/settings/image-providers/{provider}/models   { "model": "..." }
	// - DELETE /api/settings/image-providers/{provider}/models?model=...
	// - DELETE /api/settings/image-providers/{provider}
	path := strings.TrimPrefix(r.URL.Path, "/api/settings/image-providers/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	parts := strings.Split(path, "/")
	providerID := strings.ToLower(strings.TrimSpace(parts[0]))
	if providerID == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	known := map[string]bool{
		"openai_compatible": true,
		"siliconflow":       true,
		"wuyinkeji":         true,
	}
	if !known[providerID] {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unknown provider: " + providerID})
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_, err := h.st.Update(func(s *settings.Settings) error {
			if s.ImageProviders == nil {
				return nil
			}
			delete(s.ImageProviders, providerID)
			return nil
		})
		if err != nil {
			slog.Default().Warn("settings_provider_reset_failed", "provider", providerID, "err", err.Error())
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if err := h.reloadFromStore(); err != nil {
			slog.Default().Error("settings_reload_failed", "err", err.Error())
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		slog.Default().Info("settings_provider_reset", "provider", providerID)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	if len(parts) == 2 && parts[1] == "models" {
		switch r.Method {
		case http.MethodPost:
			var body struct {
				Model string `json:"model"`
			}
			if err := decodeJSON(w, r, &body); err != nil {
				return
			}
			model := strings.TrimSpace(body.Model)
			if model == "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "model is required"})
				return
			}

			baseModels := h.currentModels(providerID)
			_, err := h.st.Update(func(s *settings.Settings) error {
				if s.ImageProviders == nil {
					s.ImageProviders = map[string]settings.ProviderSettings{}
				}
				cur := s.ImageProviders[providerID]

				// If there is no explicit models override yet, start from current effective list,
				// so "add" doesn't accidentally wipe env-provided models.
				var list []string
				if cur.Models != nil {
					list = append(list, (*cur.Models)...)
				} else {
					list = append(list, baseModels...)
				}
				if !containsStr(list, model) {
					list = append(list, model)
				}
				cur.Models = &list
				s.ImageProviders[providerID] = cur
				return nil
			})
			if err != nil {
				slog.Default().Warn("settings_model_add_failed", "provider", providerID, "model", model, "err", err.Error())
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
				return
			}
			if err := h.reloadFromStore(); err != nil {
				slog.Default().Error("settings_reload_failed", "err", err.Error())
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			slog.Default().Info("settings_model_added", "provider", providerID, "model", model)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return

		case http.MethodDelete:
			model := strings.TrimSpace(r.URL.Query().Get("model"))
			if model == "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "model query param is required"})
				return
			}

			baseModels := h.currentModels(providerID)
			_, err := h.st.Update(func(s *settings.Settings) error {
				if s.ImageProviders == nil {
					s.ImageProviders = map[string]settings.ProviderSettings{}
				}
				cur := s.ImageProviders[providerID]

				var list []string
				if cur.Models != nil {
					list = append(list, (*cur.Models)...)
				} else {
					list = append(list, baseModels...)
				}
				list = removeStr(list, model)
				cur.Models = &list
				s.ImageProviders[providerID] = cur
				return nil
			})
			if err != nil {
				slog.Default().Warn("settings_model_delete_failed", "provider", providerID, "model", model, "err", err.Error())
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
				return
			}
			if err := h.reloadFromStore(); err != nil {
				slog.Default().Error("settings_reload_failed", "err", err.Error())
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			slog.Default().Info("settings_model_deleted", "provider", providerID, "model", model)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (h *Handler) currentModels(providerID string) []string {
	cfg := h.effectiveCfg()
	pc, ok := cfg.ImageProviders[providerID]
	if !ok {
		return nil
	}
	models := pc.Models
	if providerID == "wuyinkeji" {
		if provImpl, ok := h.getImageProvider("wuyinkeji"); ok {
			if p, ok := provImpl.(*wuyinkeji.Provider); ok {
				models = p.Models()
			}
		}
	}
	return append([]string{}, models...)
}

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func removeStr(list []string, v string) []string {
	var out []string
	for _, x := range list {
		if x == v {
			continue
		}
		out = append(out, x)
	}
	return out
}

func (h *Handler) effectiveCfg() config.Config {
	h.cfgMu.RLock()
	defer h.cfgMu.RUnlock()
	return h.cfg
}

func (h *Handler) getImageProvider(id string) (imageProvider, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return nil, false
	}

	// Ensure cache is up-to-date with current effective config.
	h.provMu.Lock()
	defer h.provMu.Unlock()
	h.rebuildProvidersLocked()
	p, ok := h.imageProviders[id]
	return p, ok
}

func (h *Handler) reloadFromStore() error {
	s, err := h.st.Get()
	if err != nil {
		return err
	}
	effective := runtimecfg.Merge(h.baseCfg, s)
	h.cfgMu.Lock()
	h.cfg = effective
	h.cfgMu.Unlock()

	h.provMu.Lock()
	h.imageProviders = map[string]imageProvider{}
	h.provKeys = map[string]string{}
	h.rebuildProvidersLocked()
	h.provMu.Unlock()
	slog.Default().Info("settings_reloaded")
	return nil
}

func (h *Handler) rebuildProvidersLocked() {
	// Callers must hold h.provMu.
	cfg := h.effectiveCfg()

	ensure := func(id string, key string, build func() imageProvider) {
		if prev, ok := h.provKeys[id]; ok && prev == key && h.imageProviders[id] != nil {
			return
		}
		h.imageProviders[id] = build()
		h.provKeys[id] = key
	}

	// mock is always available
	ensure("mock", "mock", func() imageProvider { return mock.New(h.staticRoot) })

	if pc, ok := cfg.ImageProviders["openai_compatible"]; ok {
		key := "openai|" + pc.BaseURL + "|" + pc.APIKey + "|" + pc.DefaultModel + "|" + strings.Join(pc.Models, ",")
		ensure("openai_compatible", key, func() imageProvider {
			return openai_compatible.New(pc.BaseURL, pc.APIKey, pc.DefaultModel, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["siliconflow"]; ok {
		key := "sf|" + pc.BaseURL + "|" + pc.APIKey + "|" + pc.DefaultModel + "|" + strings.Join(pc.Models, ",")
		ensure("siliconflow", key, func() imageProvider {
			return siliconflow.New(pc.BaseURL, pc.APIKey, pc.DefaultModel, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["wuyinkeji"]; ok {
		// if legacy mapping is used, include it in the cache key
		var kv []string
		for k, v := range pc.ModelEndpoint {
			kv = append(kv, k+"="+v)
		}
		sort.Strings(kv)
		key := "wy|" + pc.BaseURL + "|" + pc.APIKey + "|" + strings.Join(pc.Models, ",") + "|" + strings.Join(kv, ",")
		ensure("wuyinkeji", key, func() imageProvider {
			return wuyinkeji.New(pc.BaseURL, pc.APIKey, h.staticRoot, pc.Models, pc.ModelEndpoint)
		})
	}

	// Video provider (single-provider): keep existing env-driven selection, but refresh key/base from effective config.
	h.videoProv = nil
	if strings.ToLower(cfg.Provider) == "openai_compatible" || strings.ToLower(cfg.Provider) == "openai-compatible" || strings.ToLower(cfg.Provider) == "openai" {
		if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
			pc := cfg.ImageProviders["openai_compatible"]
			h.videoProv = openai_compatible.NewVideoGeneric(pc.BaseURL, pc.APIKey, cfg.VideoModel, cfg.VideoStartEP, cfg.VideoStatusEP)
		}
	}
}

func (h *Handler) videosJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.videoProv == nil {
		slog.Default().Warn("videos_create_disabled")
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "视频能力暂未启用"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	var req types.VideoJobCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	slog.Default().Info("videos_create", "provider", h.videoProv.ProviderName())
	jobID, err := h.videoProv.StartVideoJob(ctx, req)
	if err != nil {
		slog.Default().Warn("videos_create_failed", "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	now := time.Now()
	h.jobs.Put(&store.VideoJob{
		ID:              jobID,
		Status:          "queued",
		Prompt:          req.Prompt,
		DurationSeconds: req.DurationSeconds,
		AspectRatio:     req.AspectRatio,
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	writeJSON(w, http.StatusOK, types.VideoJobCreateResponse{
		JobID:    jobID,
		Status:   "queued",
		Provider: h.videoProv.ProviderName(),
	})
}

func (h *Handler) videosJobsID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.videoProv == nil {
		slog.Default().Warn("videos_get_disabled")
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "视频能力暂未启用"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/videos/jobs/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing job id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	status, videoURL, jobErr, err := h.videoProv.GetVideoJob(ctx, id)
	if err != nil {
		slog.Default().Warn("videos_get_failed", "job_id", id, "err", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	slog.Default().Info("videos_get", "job_id", id, "status", status)

	_ = h.jobs.Update(id, func(j *store.VideoJob) {
		j.Status = status
		j.VideoURL = videoURL
		j.Error = jobErr
		j.UpdatedAt = time.Now()
	})

	// If the job succeeded, persist video into MinIO and rewrite URL to this backend.
	if h.assets != nil && h.assets.Enabled() && status == "succeeded" && strings.TrimSpace(videoURL) != "" {
		job, _ := h.jobs.Get(id)
		prompt := ""
		duration := 0
		aspect := ""
		if job != nil {
			prompt = job.Prompt
			duration = job.DurationSeconds
			aspect = job.AspectRatio
		}
		a, err := h.assets.StoreRemote(ctx, assets.StoreRemoteInput{
			Capability:    "video",
			Provider:      strings.ToLower(strings.TrimSpace(h.videoProv.ProviderName())),
			Model:         strings.TrimSpace(h.effectiveCfg().VideoModel),
			Prompt:        prompt,
			ExternalJobID: id,
			Params: map[string]any{
				"duration_seconds": duration,
				"aspect_ratio":     aspect,
			},
			SourceURL: videoURL,
		})
		if err != nil {
			slog.Default().Warn("videos_store_asset_failed", "job_id", id, "err", err.Error())
		} else {
			videoURL = fmt.Sprintf("/api/assets/%d", a.ID)
			_ = h.jobs.Update(id, func(j *store.VideoJob) {
				j.VideoURL = videoURL
				j.UpdatedAt = time.Now()
			})
		}
	}

	writeJSON(w, http.StatusOK, types.VideoJobGetResponse{
		JobID:    id,
		Status:   status,
		VideoURL: videoURL,
		Error:    jobErr,
		Provider: h.videoProv.ProviderName(),
	})
}

func (h *Handler) historyList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	capability := strings.TrimSpace(r.URL.Query().Get("capability"))
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))

	items, err := h.assets.Store.List(ctx, capability, limit, offset)
	if err != nil {
		slog.Default().Warn("history_list_failed", "err", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	type row struct {
		ID          uint64 `json:"id"`
		Capability  string `json:"capability"`
		Provider    string `json:"provider"`
		Model       string `json:"model,omitempty"`
		Status      string `json:"status"`
		Error       string `json:"error,omitempty"`
		Prompt      string `json:"prompt_preview,omitempty"`
		ContentType string `json:"content_type"`
		Bytes       int64  `json:"bytes"`
		URL         string `json:"url"`
		CreatedAt   string `json:"created_at"`
	}

	out := make([]row, 0, len(items))
	for _, a := range items {
		e := ""
		if a.Error != nil {
			e = *a.Error
		}
		out = append(out, row{
			ID:          a.ID,
			Capability:  a.Capability,
			Provider:    a.Provider,
			Model:       a.Model,
			Status:      a.Status,
			Error:       e,
			Prompt:      a.PromptPreview,
			ContentType: a.ContentType,
			Bytes:       a.Bytes,
			URL:         fmt.Sprintf("/api/assets/%d", a.ID),
			CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) historyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/history/")
	id = strings.Trim(id, "/")
	if id == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	a, err := h.assets.Store.Get(ctx, uid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}
	e := ""
	if a.Error != nil {
		e = *a.Error
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":             a.ID,
		"capability":     a.Capability,
		"provider":       a.Provider,
		"model":          a.Model,
		"status":         a.Status,
		"error":          e,
		"prompt_preview": a.PromptPreview,
		"content_type":   a.ContentType,
		"bytes":          a.Bytes,
		"url":            fmt.Sprintf("/api/assets/%d", a.ID),
		"created_at":     a.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) assetsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.assets == nil || !h.assets.Enabled() || h.assets.MinIO == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	id = strings.Trim(id, "/")
	if id == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	a, err := h.assets.Store.Get(ctx, uid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	size := a.Bytes
	rangeHdr := strings.TrimSpace(r.Header.Get("Range"))
	start, end, partial := parseByteRange(rangeHdr, size)

	opts := minio.GetObjectOptions{}
	if partial {
		_ = opts.SetRange(start, end)
	}
	obj, err := h.assets.MinIO.Client.GetObject(ctx, h.assets.MinIO.Bucket, a.ObjectKey, opts)
	if err != nil {
		slog.Default().Warn("asset_get_failed", "id", uid, "err", err.Error())
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", a.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Accept-Ranges", "bytes")

	if partial && size > 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		if size > 0 {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		}
		w.WriteHeader(http.StatusOK)
	}

	if _, err := io.Copy(w, obj); err != nil {
		slog.Default().Warn("asset_stream_failed", "id", uid, "err", err.Error())
	}
}

func parseByteRange(hdr string, size int64) (start, end int64, ok bool) {
	// Only support a single range, enough for HTML5 video playback.
	if size <= 0 {
		return 0, 0, false
	}
	hdr = strings.TrimSpace(hdr)
	if hdr == "" || !strings.HasPrefix(hdr, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(hdr, "bytes=")
	// Ignore multiple ranges.
	if strings.Contains(spec, ",") {
		return 0, 0, false
	}
	a, b, found := strings.Cut(spec, "-")
	if !found {
		return 0, 0, false
	}
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	// Suffix form: bytes=-N
	if a == "" {
		n, err := strconv.ParseInt(b, 10, 64)
		if err != nil || n <= 0 {
			return 0, 0, false
		}
		if n > size {
			n = size
		}
		return size - n, size - 1, true
	}

	st, err := strconv.ParseInt(a, 10, 64)
	if err != nil || st < 0 || st >= size {
		return 0, 0, false
	}
	en := size - 1
	if b != "" {
		v, err := strconv.ParseInt(b, 10, 64)
		if err != nil || v < st {
			return 0, 0, false
		}
		if v < size {
			en = v
		}
	}
	return st, en, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// withMiddleware lives in middleware.go
