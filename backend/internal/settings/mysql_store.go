package settings

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"aigc-backend/internal/logging"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

type MySQLStore struct {
	db    *gorm.DB
	sqlDB closerPinger
	mu sync.Mutex // serialize migrations + read/modify/write updates
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
	ProviderID  string    `gorm:"column:provider_id;primaryKey;size:64"`
	Capability  string    `gorm:"column:capability;primaryKey;size:16"`
	Model       string    `gorm:"column:model;primaryKey;size:255"`
	Ord         int       `gorm:"column:ord;not null;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
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

func (st *MySQLStore) Get() (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.getLocked(context.Background())
}

func (st *MySQLStore) Update(fn func(*Settings) error) (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var out Settings
	err := st.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cur, err := st.getTx(tx)
		if err != nil {
			return err
		}
		next := deepCopy(cur)
		if err := fn(&next); err != nil {
			return err
		}
		if err := st.putTx(tx, next); err != nil {
			return err
		}
		out = next
		return nil
	})
	if err != nil {
		return Settings{}, err
	}
	return out, nil
}

func (st *MySQLStore) migrate() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS aigc_provider_configs (
			provider_id VARCHAR(64) NOT NULL,
			base_url TEXT NULL,
			base_url_set TINYINT(1) NOT NULL DEFAULT 0,
			api_key TEXT NULL,
			api_key_set TINYINT(1) NOT NULL DEFAULT 0,
			default_model VARCHAR(255) NULL,
			default_model_set TINYINT(1) NOT NULL DEFAULT 0,
			models_set TINYINT(1) NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (provider_id)
		) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS aigc_provider_models (
			provider_id VARCHAR(64) NOT NULL,
			capability VARCHAR(16) NOT NULL,
			model VARCHAR(255) NOT NULL,
			ord INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (provider_id, capability, model),
			INDEX idx_provider_capability (provider_id, capability)
		) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`,
		// Backward compatible migration (ignore error if column already exists)
		`ALTER TABLE aigc_provider_models ADD COLUMN ord INT NOT NULL DEFAULT 0`,
	}

	for _, q := range stmts {
		if err := st.db.WithContext(ctx).Exec(q).Error; err != nil {
			// Some migrations are "best effort" for existing installations.
			// Example: adding a column that may already exist.
			if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(q)), "ALTER TABLE") {
				continue
			}
			return err
		}
	}
	slog.Default().Info("settings_store_mysql_migrated")
	return nil
}

func (st *MySQLStore) getLocked(ctx context.Context) (Settings, error) {
	tx := st.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Settings{}, tx.Error
	}
	s, err := st.getTx(tx)
	if err != nil {
		_ = tx.Rollback().Error
		return Settings{}, err
	}
	if err := tx.Commit().Error; err != nil {
		return Settings{}, err
	}
	return s, nil
}

func (st *MySQLStore) getTx(tx *gorm.DB) (Settings, error) {
	s := Settings{ImageProviders: map[string]ProviderSettings{}}

	var cfgs []providerConfigRow
	if err := tx.Find(&cfgs).Error; err != nil {
		return Settings{}, err
	}

	needModels := map[string]bool{}
	for _, r := range cfgs {
		id := strings.ToLower(strings.TrimSpace(r.ProviderID))
		if id == "" {
			continue
		}
		ps := ProviderSettings{}
		if r.BaseURLSet {
			v := ""
			if r.BaseURL != nil {
				v = *r.BaseURL
			}
			ps.BaseURL = &v
		}
		if r.APIKeySet {
			v := ""
			if r.APIKey != nil {
				v = *r.APIKey
			}
			ps.APIKey = &v
		}
		if r.DefaultModelSet {
			v := ""
			if r.DefaultModel != nil {
				v = *r.DefaultModel
			}
			ps.DefaultModel = &v
		}
		if r.ModelsSet {
			needModels[id] = true
			empty := []string{}
			ps.Models = &empty
		}
		// only keep providers with at least one override field set
		if ps.BaseURL != nil || ps.APIKey != nil || ps.DefaultModel != nil || ps.Models != nil {
			s.ImageProviders[id] = ps
		}
	}

	if len(needModels) == 0 {
		if len(s.ImageProviders) == 0 {
			return Settings{}, nil
		}
		return s, nil
	}

	var mrows []providerModelRow
	if err := tx.Where("capability = ?", "image").Order("provider_id ASC, ord ASC, created_at ASC").Find(&mrows).Error; err != nil {
		return Settings{}, err
	}
	byProv := map[string][]string{}
	for _, r := range mrows {
		pid := strings.ToLower(strings.TrimSpace(r.ProviderID))
		if pid == "" {
			continue
		}
		byProv[pid] = append(byProv[pid], strings.TrimSpace(r.Model))
	}

	for pid := range needModels {
		ps, ok := s.ImageProviders[pid]
		if !ok || ps.Models == nil {
			continue
		}
		list := dedupeKeepOrder(byProv[pid])
		ps.Models = &list
		s.ImageProviders[pid] = ps
	}

	return s, nil
}

func (st *MySQLStore) putTx(tx *gorm.DB, s Settings) error {
	desired := map[string]ProviderSettings{}
	if s.ImageProviders != nil {
		for k, v := range s.ImageProviders {
			desired[strings.ToLower(strings.TrimSpace(k))] = v
		}
	}

	// Delete providers that were removed from settings (or are now empty).
	var existing []providerConfigRow
	if err := tx.Select("provider_id").Find(&existing).Error; err != nil {
		return err
	}
	for _, r := range existing {
		id := strings.ToLower(strings.TrimSpace(r.ProviderID))
		ps, ok := desired[id]
		if !ok || (ps.BaseURL == nil && ps.APIKey == nil && ps.DefaultModel == nil && ps.Models == nil) {
			if err := tx.Where("provider_id = ?", id).Delete(&providerConfigRow{}).Error; err != nil {
				return err
			}
			if err := tx.Where("provider_id = ? AND capability = ?", id, "image").Delete(&providerModelRow{}).Error; err != nil {
				return err
			}
			delete(desired, id)
		}
	}

	for pid, ps := range desired {
		pid = strings.ToLower(strings.TrimSpace(pid))
		if pid == "" {
			continue
		}

		baseURLSet := ps.BaseURL != nil
		apiKeySet := ps.APIKey != nil
		defModelSet := ps.DefaultModel != nil
		modelsSet := ps.Models != nil

		cfg := providerConfigRow{
			ProviderID:      pid,
			BaseURLSet:      baseURLSet,
			APIKeySet:       apiKeySet,
			DefaultModelSet: defModelSet,
			ModelsSet:       modelsSet,
		}
		if baseURLSet {
			v := strings.TrimSpace(*ps.BaseURL)
			cfg.BaseURL = &v
		}
		if apiKeySet {
			v := strings.TrimSpace(*ps.APIKey)
			cfg.APIKey = &v
		}
		if defModelSet {
			v := strings.TrimSpace(*ps.DefaultModel)
			cfg.DefaultModel = &v
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "provider_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"base_url", "base_url_set", "api_key", "api_key_set", "default_model", "default_model_set", "models_set"}),
		}).Create(&cfg).Error; err != nil {
			return err
		}

		// models: always sync when models_set is known (true/false) in settings.
		// If models is nil, we consider it "inherit", and thus delete stored models.
		if err := tx.Where("provider_id = ? AND capability = ?", pid, "image").Delete(&providerModelRow{}).Error; err != nil {
			return err
		}
		if modelsSet {
			models := dedupeKeepOrder(*ps.Models)
			var batch []providerModelRow
			for i, m := range models {
				if m == "" {
					continue
				}
				if len(m) > 255 {
					return fmt.Errorf("model too long for provider %s", pid)
				}
				batch = append(batch, providerModelRow{
					ProviderID: pid,
					Capability: "image",
					Model:      m,
					Ord:        i,
				})
			}
			if len(batch) > 0 {
				if err := tx.Create(&batch).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func dedupeKeepOrder(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range in {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		if seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x)
	}
	return out
}
