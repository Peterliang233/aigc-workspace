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

	// Video providers (multi-provider). IDs are stable and match frontend selection.
	if h.videoProviders == nil {
		h.videoProviders = map[string]videoProvider{}
	}
	if h.videoProvKeys == nil {
		h.videoProvKeys = map[string]string{}
	}
	ensureV := func(id string, key string, build func() videoProvider) {
		if prev, ok := h.videoProvKeys[id]; ok && prev == key && h.videoProviders[id] != nil {
			return
		}
		h.videoProviders[id] = build()
		h.videoProvKeys[id] = key
	}

	if pc, ok := cfg.ImageProviders["siliconflow"]; ok && strings.TrimSpace(pc.APIKey) != "" {
		key := "sfv|" + pc.BaseURL + "|" + pc.APIKey
		ensureV("siliconflow", key, func() videoProvider {
			return siliconflow.NewVideo(pc.BaseURL, pc.APIKey)
		})
	}

	// Generic async video API (env provides endpoints). Binds to openai_compatible credentials.
	if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
		pc := cfg.ImageProviders["openai_compatible"]
		if strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
			key := "ov|" + pc.BaseURL + "|" + pc.APIKey + "|" + cfg.VideoModel + "|" + cfg.VideoStartEP + "|" + cfg.VideoStatusEP
			ensureV("openai_compatible", key, func() videoProvider {
				return openai_compatible.NewVideoGeneric(pc.BaseURL, pc.APIKey, cfg.VideoModel, cfg.VideoStartEP, cfg.VideoStatusEP)
			})
		}
	}
}
