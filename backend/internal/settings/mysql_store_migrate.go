package settings

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

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
			if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(q)), "ALTER TABLE") {
				continue
			}
			return err
		}
	}
	slog.Default().Info("settings_store_mysql_migrated")
	return nil
}

