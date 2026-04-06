package jeniya

import (
	"context"

	"aigc-backend/internal/types"
)

func (p *Provider) GenerateAudio(ctx context.Context, req types.AudioGenerateRequest) (types.AudioGenerateResponse, error) {
	resp, err := p.audio.GenerateAudio(ctx, req)
	if err != nil {
		return types.AudioGenerateResponse{}, err
	}
	resp.Provider = p.ProviderName()
	return resp, nil
}
