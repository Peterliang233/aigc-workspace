package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"aigc-backend/internal/logging"
)

type MySQLStore struct {
	db *sql.DB
	mu sync.Mutex // serialize migrations + read/modify/write updates
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("MYSQL_DSN is empty")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	st := &MySQLStore{db: db}
	if err := st.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	slog.Default().Info("settings_store_mysql_ready", "dsn", logging.RedactDSN(dsn))
	return st, nil
}

func (st *MySQLStore) Close() error {
	if st.db == nil {
		return nil
	}
	return st.db.Close()
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

	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return Settings{}, err
	}
	defer tx.Rollback()

	cur, err := st.getTx(ctx, tx)
	if err != nil {
		return Settings{}, err
	}
	next := deepCopy(cur)
	if err := fn(&next); err != nil {
		return Settings{}, err
	}

	if err := st.putTx(ctx, tx, next); err != nil {
		return Settings{}, err
	}
	if err := tx.Commit(); err != nil {
		return Settings{}, err
	}
	return next, nil
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
		if _, err := st.db.ExecContext(ctx, q); err != nil {
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
	tx, err := st.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Settings{}, err
	}
	defer tx.Rollback()

	s, err := st.getTx(ctx, tx)
	if err != nil {
		return Settings{}, err
	}
	_ = tx.Commit()
	return s, nil
}

func (st *MySQLStore) getTx(ctx context.Context, tx *sql.Tx) (Settings, error) {
	s := Settings{ImageProviders: map[string]ProviderSettings{}}

	rows, err := tx.QueryContext(ctx, `SELECT provider_id, base_url, base_url_set, api_key, api_key_set, default_model, default_model_set, models_set FROM aigc_provider_configs`)
	if err != nil {
		return Settings{}, err
	}
	defer rows.Close()

	type cfgRow struct {
		ProviderID      string
		BaseURL         sql.NullString
		BaseURLSet      bool
		APIKey          sql.NullString
		APIKeySet       bool
		DefaultModel    sql.NullString
		DefaultModelSet bool
		ModelsSet       bool
	}
	var cfgs []cfgRow
	for rows.Next() {
		var r cfgRow
		if err := rows.Scan(&r.ProviderID, &r.BaseURL, &r.BaseURLSet, &r.APIKey, &r.APIKeySet, &r.DefaultModel, &r.DefaultModelSet, &r.ModelsSet); err != nil {
			return Settings{}, err
		}
		cfgs = append(cfgs, r)
	}
	if err := rows.Err(); err != nil {
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
			if r.BaseURL.Valid {
				v = r.BaseURL.String
			}
			ps.BaseURL = &v
		}
		if r.APIKeySet {
			v := ""
			if r.APIKey.Valid {
				v = r.APIKey.String
			}
			ps.APIKey = &v
		}
		if r.DefaultModelSet {
			v := ""
			if r.DefaultModel.Valid {
				v = r.DefaultModel.String
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

	mrows, err := tx.QueryContext(ctx, `SELECT provider_id, model FROM aigc_provider_models WHERE capability='image' ORDER BY provider_id ASC, ord ASC, created_at ASC`)
	if err != nil {
		return Settings{}, err
	}
	defer mrows.Close()

	byProv := map[string][]string{}
	for mrows.Next() {
		var pid, model string
		if err := mrows.Scan(&pid, &model); err != nil {
			return Settings{}, err
		}
		pid = strings.ToLower(strings.TrimSpace(pid))
		if pid == "" {
			continue
		}
		byProv[pid] = append(byProv[pid], strings.TrimSpace(model))
	}
	if err := mrows.Err(); err != nil {
		return Settings{}, err
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

func (st *MySQLStore) putTx(ctx context.Context, tx *sql.Tx, s Settings) error {
	if s.ImageProviders == nil {
		return nil
	}

	for pid, ps := range s.ImageProviders {
		pid = strings.ToLower(strings.TrimSpace(pid))
		if pid == "" {
			continue
		}

		baseURLSet := ps.BaseURL != nil
		apiKeySet := ps.APIKey != nil
		defModelSet := ps.DefaultModel != nil
		modelsSet := ps.Models != nil

		var baseURL sql.NullString
		if baseURLSet {
			baseURL = sql.NullString{String: strings.TrimSpace(*ps.BaseURL), Valid: true}
		}
		var apiKey sql.NullString
		if apiKeySet {
			apiKey = sql.NullString{String: strings.TrimSpace(*ps.APIKey), Valid: true}
		}
		var defModel sql.NullString
		if defModelSet {
			defModel = sql.NullString{String: strings.TrimSpace(*ps.DefaultModel), Valid: true}
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO aigc_provider_configs
				(provider_id, base_url, base_url_set, api_key, api_key_set, default_model, default_model_set, models_set)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				base_url=VALUES(base_url),
				base_url_set=VALUES(base_url_set),
				api_key=VALUES(api_key),
				api_key_set=VALUES(api_key_set),
				default_model=VALUES(default_model),
				default_model_set=VALUES(default_model_set),
				models_set=VALUES(models_set)
		`, pid, nullOrString(baseURLSet, baseURL), boolToInt(baseURLSet), nullOrString(apiKeySet, apiKey), boolToInt(apiKeySet), nullOrString(defModelSet, defModel), boolToInt(defModelSet), boolToInt(modelsSet))
		if err != nil {
			return err
		}

		// models: always sync when models_set is known (true/false) in settings.
		// If models is nil, we consider it "inherit", and thus delete stored models.
		if _, err := tx.ExecContext(ctx, `DELETE FROM aigc_provider_models WHERE provider_id=? AND capability='image'`, pid); err != nil {
			return err
		}
		if modelsSet {
			models := dedupeKeepOrder(*ps.Models)
			for i, m := range models {
				if m == "" {
					continue
				}
				if len(m) > 255 {
					return fmt.Errorf("model too long for provider %s", pid)
				}
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO aigc_provider_models (provider_id, capability, model, ord)
					VALUES (?, 'image', ?, ?)
					ON DUPLICATE KEY UPDATE ord=VALUES(ord)
				`, pid, m, i); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func nullOrString(set bool, v sql.NullString) any {
	if !set {
		return nil
	}
	if v.Valid {
		return v.String
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
