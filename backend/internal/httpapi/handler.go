package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/config"
	"aigc-backend/internal/providers/mock"
	"aigc-backend/internal/providers/openai_compatible"
	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

type Handler struct {
	cfg config.Config

	imageProv interface {
		ProviderName() string
		GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error)
	}
	videoProv interface {
		ProviderName() string
		StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error)
		GetVideoJob(ctx context.Context, jobID string) (string, string, string, error)
	}

	jobs       *store.JobStore
	staticRoot string
}

func NewHandler(cfg config.Config) http.Handler {
	staticRoot := filepath.Join("var")
	_ = os.MkdirAll(filepath.Join(staticRoot, "generated"), 0o755)

	h := &Handler{
		cfg:        cfg,
		jobs:       store.NewJobStore(),
		staticRoot: staticRoot,
	}

	switch strings.ToLower(cfg.Provider) {
	case "openai_compatible", "openai-compatible", "openai":
		h.imageProv = openai_compatible.New(cfg.BaseURL, cfg.APIKey, cfg.ImageModel, staticRoot)
		if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
			h.videoProv = openai_compatible.NewVideoGeneric(cfg.BaseURL, cfg.APIKey, cfg.VideoModel, cfg.VideoStartEP, cfg.VideoStatusEP)
		}
	default:
		h.imageProv = mock.New(staticRoot)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.healthz)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticRoot))))
	mux.HandleFunc("/api/images/generate", h.imagesGenerate)
	mux.HandleFunc("/api/videos/jobs", h.videosJobs)
	mux.HandleFunc("/api/videos/jobs/", h.videosJobsID) // GET /api/videos/jobs/{id}

	return withMiddleware(mux, cfg.AllowedOrigins)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"provider": h.imageProv.ProviderName(),
	})
}

func (h *Handler) imagesGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	var req types.ImageGenerateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.imageProv.GenerateImage(ctx, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	// Ensure provider field present even if provider implementation left it blank.
	if strings.TrimSpace(resp.Provider) == "" {
		resp.Provider = h.imageProv.ProviderName()
	}
	writeJSON(w, http.StatusOK, resp)
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
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "video provider not configured; set AIGC_PROVIDER=openai_compatible and configure AIGC_VIDEO_* endpoints"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req types.VideoJobCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	jobID, err := h.videoProv.StartVideoJob(ctx, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	now := time.Now()
	h.jobs.Put(&store.VideoJob{
		ID:        jobID,
		Status:    "queued",
		CreatedAt: now,
		UpdatedAt: now,
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
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "video provider not configured"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/videos/jobs/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing job id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	status, videoURL, jobErr, err := h.videoProv.GetVideoJob(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	_ = h.jobs.Update(id, func(j *store.VideoJob) {
		j.Status = status
		j.VideoURL = videoURL
		j.Error = jobErr
		j.UpdatedAt = time.Now()
	})

	writeJSON(w, http.StatusOK, types.VideoJobGetResponse{
		JobID:    id,
		Status:   status,
		VideoURL: videoURL,
		Error:    jobErr,
		Provider: h.videoProv.ProviderName(),
	})
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

func withMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := map[string]bool{}
	for _, o := range allowedOrigins {
		allowed[strings.TrimSpace(o)] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
