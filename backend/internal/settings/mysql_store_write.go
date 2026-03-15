package settings

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

