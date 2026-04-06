package jeniya

import (
	"context"

	"aigc-backend/internal/types"
)

func (p *Provider) GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error) {
	resp, err := p.image.GenerateImage(ctx, req)
	if err != nil {
		return types.ImageGenerateResponse{}, err
	}
	resp.Provider = p.ProviderName()
	return resp, nil
}
