package storyvideo

import (
	"context"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func NewStore(dsn string) (*Store, error) {
	dsn = strings.TrimSpace(dsn)
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	st := &Store{db: gdb}
	return st, st.Migrate()
}

func (s *Store) Migrate() error {
	return s.db.AutoMigrate(&Project{}, &Shot{}, &Event{})
}

func (s *Store) CreateProject(ctx context.Context, project *Project, shots []Shot) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(project).Error; err != nil {
			return err
		}
		if len(shots) == 0 {
			return nil
		}
		return tx.Create(&shots).Error
	})
}

func (s *Store) ReplaceDraft(ctx context.Context, project *Project, shots []Shot) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Project{}).Where("id = ?", project.ID).Updates(project).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", project.ID).Delete(&Shot{}).Error; err != nil {
			return err
		}
		if len(shots) == 0 {
			return nil
		}
		return tx.Create(&shots).Error
	})
}

func (s *Store) GetProject(ctx context.Context, id string) (*Project, []Shot, error) {
	var project Project
	if err := s.db.WithContext(ctx).First(&project, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	var shots []Shot
	if err := s.db.WithContext(ctx).Where("project_id = ?", id).Order("shot_index ASC").Find(&shots).Error; err != nil {
		return nil, nil, err
	}
	return &project, shots, nil
}

func (s *Store) ListProjects(ctx context.Context, limit int) ([]Project, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var out []Project
	err := s.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (s *Store) UpdateProject(ctx context.Context, id string, attrs map[string]any) error {
	return s.db.WithContext(ctx).Model(&Project{}).Where("id = ?", id).Updates(attrs).Error
}

func (s *Store) UpdateShot(ctx context.Context, id string, attrs map[string]any) error {
	return s.db.WithContext(ctx).Model(&Shot{}).Where("id = ?", id).Updates(attrs).Error
}
