package httpapi

import (
	"sort"
	"strings"

	"aigc-backend/internal/providers/mock"
	"aigc-backend/internal/providers/openai_compatible"
	"aigc-backend/internal/providers/siliconflow"
	"aigc-backend/internal/providers/wuyinkeji"
)

func (h *Handler) rebuildProvidersLocked() {
	// Callers must hold h.provMu.
	cfg := h.effectiveCfg()

	ensure := func(id string, key string, build func() imageProvider) {
		if prev, ok := h.provKeys[id]; ok && prev == key && h.imageProviders[id] != nil {
			return
		}
		h.imageProviders[id] = build()
		h.provKeys[id] = key
	}

	// mock is always available
	ensure("mock", "mock", func() imageProvider { return mock.New(h.staticRoot) })

	if pc, ok := cfg.ImageProviders["openai_compatible"]; ok {
		key := "openai|" + pc.BaseURL + "|" + pc.APIKey + "|" + pc.DefaultModel + "|" + strings.Join(pc.Models, ",")
		ensure("openai_compatible", key, func() imageProvider {
			return openai_compatible.New(pc.BaseURL, pc.APIKey, pc.DefaultModel, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["siliconflow"]; ok {
		key := "sf|" + pc.BaseURL + "|" + pc.APIKey + "|" + pc.DefaultModel + "|" + strings.Join(pc.Models, ",")
		ensure("siliconflow", key, func() imageProvider {
			return siliconflow.New(pc.BaseURL, pc.APIKey, pc.DefaultModel, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["wuyinkeji"]; ok {
		// if legacy mapping is used, include it in the cache key
		var kv []string
		for k, v := range pc.ModelEndpoint {
			kv = append(kv, k+"="+v)
		}
		sort.Strings(kv)
		key := "wy|" + pc.BaseURL + "|" + pc.APIKey + "|" + strings.Join(pc.Models, ",") + "|" + strings.Join(kv, ",")
		ensure("wuyinkeji", key, func() imageProvider {
			return wuyinkeji.New(pc.BaseURL, pc.APIKey, h.staticRoot, pc.Models, pc.ModelEndpoint)
		})
	}

	// Video provider (single-provider): keep existing env-driven selection.
	h.videoProv = nil
	if strings.ToLower(cfg.Provider) == "openai_compatible" || strings.ToLower(cfg.Provider) == "openai-compatible" || strings.ToLower(cfg.Provider) == "openai" {
		if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
			pc := cfg.ImageProviders["openai_compatible"]
			h.videoProv = openai_compatible.NewVideoGeneric(pc.BaseURL, pc.APIKey, cfg.VideoModel, cfg.VideoStartEP, cfg.VideoStatusEP)
		}
	}
}

