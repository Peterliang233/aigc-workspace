package runtimecfg

import (
	"strings"

	"aigc-backend/internal/config"
	"aigc-backend/internal/settings"
)

// Merge overlays persisted settings on top of base env config.
// Settings fields that are nil are treated as "inherit".
func Merge(base config.Config, s settings.Settings) config.Config {
	out := base

	if out.ImageProviders == nil {
		out.ImageProviders = map[string]config.ImageProviderConfig{}
	}

	for pid, ps := range s.ImageProviders {
		pid = strings.ToLower(strings.TrimSpace(pid))
		if pid == "" {
			continue
		}
		pc := out.ImageProviders[pid]

		// Security and deploy-time concerns:
		// - Base URL / API Key / DefaultModel are managed via environment variables, not via runtime settings.
		//   This avoids persisting secrets in DB and makes deployments reproducible.
		// - We only allow model list to be managed at runtime (add/remove).
		if ps.Models != nil {
			// preserve order; UI controls it
			var models []string
			for _, m := range *ps.Models {
				m = strings.TrimSpace(m)
				if m != "" {
					models = append(models, m)
				}
			}
			pc.Models = models
		}

		out.ImageProviders[pid] = pc
	}

	return out
}
