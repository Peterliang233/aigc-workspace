package settings

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"aigc-backend/internal/logging"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type MySQLStore struct {
	db    *gorm.DB
	sqlDB closerPinger
	mu    sync.Mutex // serialize migrations + read/modify/write updates
}

type closerPinger interface {
	Close() error
	PingContext(ctx context.Context) error
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d time.Duration)
}

type providerConfigRow struct {
	ProviderID      string     `gorm:"column:provider_id;primaryKey;size:64"`
	BaseURL         *string    `gorm:"column:base_url;type:text"`
	BaseURLSet      bool       `gorm:"column:base_url_set;not null;default:0"`
	APIKey          *string    `gorm:"column:api_key;type:text"`
	APIKeySet       bool       `gorm:"column:api_key_set;not null;default:0"`
	DefaultModel    *string    `gorm:"column:default_model;size:255"`
	DefaultModelSet bool       `gorm:"column:default_model_set;not null;default:0"`
	ModelsSet       bool       `gorm:"column:models_set;not null;default:0"`
	UpdatedAt       *time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (providerConfigRow) TableName() string { return "aigc_provider_configs" }

type providerModelRow struct {
	ProviderID string    `gorm:"column:provider_id;primaryKey;size:64"`
	Capability string    `gorm:"column:capability;primaryKey;size:16"`
	Model      string    `gorm:"column:model;primaryKey;size:255"`
	Ord        int       `gorm:"column:ord;not null;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (providerModelRow) TableName() string { return "aigc_provider_models" }

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("MYSQL_DSN is empty")
	}

	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Avoid noisy stdout logging; we log high-level events ourselves via slog.
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	st := &MySQLStore{db: gdb, sqlDB: sqlDB}
	if err := st.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	slog.Default().Info("settings_store_mysql_ready", "dsn", logging.RedactDSN(dsn))
	return st, nil
}

func (st *MySQLStore) Close() error {
	if st.sqlDB == nil {
		return nil
	}
	return st.sqlDB.Close()
}

