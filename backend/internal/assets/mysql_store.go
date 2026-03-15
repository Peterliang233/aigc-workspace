package assets

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"aigc-backend/internal/logging"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type MySQLStore struct {
	db    *gorm.DB
	sqlDB closerPinger
}

type closerPinger interface {
	Close() error
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d time.Duration)
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("MYSQL_DSN is empty")
	}

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

	st := &MySQLStore{db: gdb, sqlDB: sqlDB}
	if err := st.Migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	slog.Default().Info("assets_store_mysql_ready", "dsn", logging.RedactDSN(dsn))
	return st, nil
}

func (st *MySQLStore) Close() error {
	if st == nil || st.sqlDB == nil {
		return nil
	}
	return st.sqlDB.Close()
}

func (st *MySQLStore) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS aigc_generation_assets (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			capability VARCHAR(16) NOT NULL,
			provider VARCHAR(64) NOT NULL,
			model VARCHAR(255) NOT NULL DEFAULT '',
			prompt_sha256 CHAR(64) NOT NULL DEFAULT '',
			prompt_preview VARCHAR(255) NOT NULL DEFAULT '',
			params_json JSON NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'succeeded',
			error TEXT NULL,
			source_url TEXT NULL,
			object_key VARCHAR(512) NOT NULL,
			content_type VARCHAR(128) NOT NULL DEFAULT 'application/octet-stream',
			bytes BIGINT NOT NULL DEFAULT 0,
			external_job_id VARCHAR(128) NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_external_job_id (external_job_id),
			INDEX idx_cap_created (capability, created_at)
		) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`,
	}
	for _, q := range stmts {
		if err := st.db.WithContext(ctx).Exec(q).Error; err != nil {
			return err
		}
	}
	slog.Default().Info("assets_store_mysql_migrated")
	return nil
}

func (st *MySQLStore) Create(ctx context.Context, a *Asset) (*Asset, error) {
	if err := st.db.WithContext(ctx).Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (st *MySQLStore) Get(ctx context.Context, id uint64) (*Asset, error) {
	var a Asset
	if err := st.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (st *MySQLStore) FindByExternalJobID(ctx context.Context, jobID string) (*Asset, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, nil
	}
	var a Asset
	if err := st.db.WithContext(ctx).First(&a, "external_job_id = ?", jobID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (st *MySQLStore) List(ctx context.Context, capability string, limit, offset int) ([]Asset, error) {
	capability = strings.ToLower(strings.TrimSpace(capability))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	q := st.db.WithContext(ctx).Model(&Asset{}).Order("id DESC").Limit(limit).Offset(offset)
	if capability != "" && capability != "all" {
		q = q.Where("capability = ?", capability)
	}
	var out []Asset
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
