package providers

import (
	"context"

	"aigc-backend/internal/types"
)

type ImageProvider interface {
	ProviderName() string
	GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error)
}

type VideoProvider interface {
	ProviderName() string
	StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (jobID string, err error)
	GetVideoJob(ctx context.Context, jobID string) (status string, videoURL string, jobErr string, err error)
}
