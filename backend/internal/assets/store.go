package assets

import "context"

type Store interface {
	Close() error
	Migrate() error

	Create(ctx context.Context, a *Asset) (*Asset, error)
	Get(ctx context.Context, id uint64) (*Asset, error)
	FindByExternalJobID(ctx context.Context, jobID string) (*Asset, error)
	List(ctx context.Context, capability string, limit, offset int) ([]Asset, error)
}
