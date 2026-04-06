package httpapi

import (
	"context"
	"sync"

	"aigc-backend/internal/animation"
	"aigc-backend/internal/assets"
	"aigc-backend/internal/config"
	"aigc-backend/internal/modelcfg"
	"aigc-backend/internal/store"
	"aigc-backend/internal/types"
)

type Handler struct {
	cfg    config.Config
	models *modelcfg.Config
	assets *assets.Service

	provMu         sync.Mutex
	imageProviders map[string]imageProvider
	provKeys       map[string]string
	videoProviders map[string]videoProvider
	videoProvKeys  map[string]string
	audioProviders map[string]audioProvider
	audioProvKeys  map[string]string

	jobs          *store.JobStore
	animationJobs *store.AnimationStore
	mediaWorker   *animation.MediaClient
	staticRoot    string
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

type audioProvider interface {
	ProviderName() string
	GenerateAudio(ctx context.Context, req types.AudioGenerateRequest) (types.AudioGenerateResponse, error)
}
