package assets

import "context"

type ListOptions struct {
	Capability string
	Query      string
	Limit      int
	Offset     int
}

type Store interface {
	Close() error
	Migrate() error

	Create(ctx context.Context, a *Asset) (*Asset, error)
	Get(ctx context.Context, id uint64) (*Asset, error)
	Delete(ctx context.Context, id uint64) error
	FindByExternalJobID(ctx context.Context, jobID string) (*Asset, error)
	List(ctx context.Context, opt ListOptions) ([]Asset, int64, error)
}
