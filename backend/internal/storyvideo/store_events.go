package storyvideo

import (
	"context"

	"gorm.io/gorm"
)

func (s *Store) GetShot(ctx context.Context, id string) (*Shot, error) {
	var shot Shot
	if err := s.db.WithContext(ctx).First(&shot, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &shot, nil
}

func (s *Store) ListEvents(ctx context.Context, projectID string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var out []Event
	err := s.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

func (s *Store) AddEvent(ctx context.Context, event *Event) error {
	return s.db.WithContext(ctx).Create(event).Error
}

func (s *Store) WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(fn)
}
