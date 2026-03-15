package settings

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

func (st *MySQLStore) Get() (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.getLocked(context.Background())
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
