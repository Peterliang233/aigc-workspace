package modelcfg

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

func Load(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "models.json"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Version <= 0 {
		return nil, errors.New("models config: missing/invalid version")
	}
	for i := range cfg.Providers {
		cfg.Providers[i].ID = strings.ToLower(strings.TrimSpace(cfg.Providers[i].ID))
		cfg.Providers[i].Label = strings.TrimSpace(cfg.Providers[i].Label)
	}
	cfg.Defaults.ImageProvider = strings.ToLower(strings.TrimSpace(cfg.Defaults.ImageProvider))
	cfg.Defaults.VideoProvider = strings.ToLower(strings.TrimSpace(cfg.Defaults.VideoProvider))
	return &cfg, nil
}
