package httpapi

import (
	"strings"

	"aigc-backend/internal/providers/gptbest"
	"aigc-backend/internal/providers/jeniya"
	"aigc-backend/internal/providers/mock"
	"aigc-backend/internal/providers/openai_compatible"
	"aigc-backend/internal/providers/siliconflow"
	"aigc-backend/internal/providers/wuyinkeji"
)

func (h *Handler) rebuildProvidersLocked() {
	// Callers must hold h.provMu.
	cfg := h.cfg
	defImageModel := func(providerID string) string {
		if h.models == nil {
			return ""
		}
		return h.models.DefaultModel(providerID, "image")
	}

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
		dm := defImageModel("openai_compatible")
		key := "openai|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensure("openai_compatible", key, func() imageProvider {
			return openai_compatible.New(pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["siliconflow"]; ok {
		dm := defImageModel("siliconflow")
		key := "sf|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensure("siliconflow", key, func() imageProvider {
			return siliconflow.New(pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["wuyinkeji"]; ok {
		key := "wy|" + pc.BaseURL + "|" + pc.APIKey
		ensure("wuyinkeji", key, func() imageProvider {
			return wuyinkeji.New(pc.BaseURL, pc.APIKey, h.staticRoot, nil, nil)
		})
	}
	if pc, ok := cfg.ImageProviders["jeniya"]; ok {
		dm := defImageModel("jeniya")
		key := "jeniya|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensure("jeniya", key, func() imageProvider {
			return jeniya.New(pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["bltcy"]; ok {
		dm := defImageModel("bltcy")
		key := "bltcy|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensure("bltcy", key, func() imageProvider {
			return gptbest.New("bltcy", pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["gpt_best"]; ok {
		dm := defImageModel("gpt_best")
		key := "gptbest|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensure("gpt_best", key, func() imageProvider {
			return gptbest.New("gpt_best", pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}

	if h.audioProviders == nil {
		h.audioProviders = map[string]audioProvider{}
	}
	if h.audioProvKeys == nil {
		h.audioProvKeys = map[string]string{}
	}
	ensureA := func(id string, key string, build func() audioProvider) {
		if prev, ok := h.audioProvKeys[id]; ok && prev == key && h.audioProviders[id] != nil {
			return
		}
		h.audioProviders[id] = build()
		h.audioProvKeys[id] = key
	}
	if pc, ok := cfg.ImageProviders["bltcy"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		dm := ""
		if h.models != nil {
			dm = h.models.DefaultModel("bltcy", "audio")
		}
		key := "bltcya|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensureA("bltcy", key, func() audioProvider {
			return gptbest.New("bltcy", pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["wuyinkeji"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		key := "wya|" + pc.BaseURL + "|" + pc.APIKey
		ensureA("wuyinkeji", key, func() audioProvider {
			return wuyinkeji.New(pc.BaseURL, pc.APIKey, h.staticRoot, nil, nil)
		})
	}
	if pc, ok := cfg.ImageProviders["jeniya"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		dm := ""
		if h.models != nil {
			dm = h.models.DefaultModel("jeniya", "audio")
		}
		key := "jeniyaa|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensureA("jeniya", key, func() audioProvider {
			return jeniya.New(pc.BaseURL, pc.APIKey, dm, h.staticRoot)
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
	if pc, ok := cfg.ImageProviders["wuyinkeji"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		key := "wyv|" + pc.BaseURL + "|" + pc.APIKey
		ensureV("wuyinkeji", key, func() videoProvider {
			return wuyinkeji.New(pc.BaseURL, pc.APIKey, h.staticRoot, nil, nil)
		})
	}
	if pc, ok := cfg.ImageProviders["bltcy"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		dm := ""
		if h.models != nil {
			dm = h.models.DefaultModel("bltcy", "image")
		}
		key := "bltcyv|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensureV("bltcy", key, func() videoProvider {
			return gptbest.New("bltcy", pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["jeniya"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		dm := ""
		if h.models != nil {
			dm = h.models.DefaultModel("jeniya", "image")
		}
		key := "jeniyav|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensureV("jeniya", key, func() videoProvider {
			return jeniya.New(pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}
	if pc, ok := cfg.ImageProviders["gpt_best"]; ok && strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
		dm := ""
		if h.models != nil {
			dm = h.models.DefaultModel("gpt_best", "image")
		}
		key := "gptbestv|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm
		ensureV("gpt_best", key, func() videoProvider {
			return gptbest.New("gpt_best", pc.BaseURL, pc.APIKey, dm, h.staticRoot)
		})
	}

	// Generic async video API (env provides endpoints). Binds to openai_compatible credentials.
	if cfg.VideoStartEP != "" && cfg.VideoStatusEP != "" {
		pc := cfg.ImageProviders["openai_compatible"]
		if strings.TrimSpace(pc.BaseURL) != "" && strings.TrimSpace(pc.APIKey) != "" {
			dm := ""
			if h.models != nil {
				dm = h.models.DefaultModel("openai_compatible", "video")
			}
			key := "ov|" + pc.BaseURL + "|" + pc.APIKey + "|" + dm + "|" + cfg.VideoStartEP + "|" + cfg.VideoStatusEP
			ensureV("openai_compatible", key, func() videoProvider {
				return openai_compatible.NewVideoGeneric(pc.BaseURL, pc.APIKey, dm, cfg.VideoStartEP, cfg.VideoStatusEP)
			})
		}
	}
}
