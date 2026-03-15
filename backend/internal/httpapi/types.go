package httpapi

import (
	"context"
	"sync"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/config"
	"aigc-backend/internal/settings"
	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
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
	videoProv      videoProvider

	jobs       *store.JobStore
	staticRoot string
}

type imageProvider interface {
	ProviderName() string
	GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error)
}

type videoProvider interface {
	ProviderName() string
	StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error)
	GetVideoJob(ctx context.Context, jobID string) (string, string, string, error)
}

