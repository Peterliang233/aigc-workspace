package modelcfg

import "strings"

func (c *Config) Provider(id string) *Provider {
	if c == nil {
		return nil
	}
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return nil
	}
	for i := range c.Providers {
		if c.Providers[i].ID == id {
			return &c.Providers[i]
		}
	}
	return nil
}

func (c *Config) DefaultProvider(capability string) string {
	if c == nil {
		return ""
	}
	capability = strings.ToLower(strings.TrimSpace(capability))
	if capability == "video" {
		return c.Defaults.VideoProvider
	}
	if capability == "audio" {
		return c.Defaults.AudioProvider
	}
	return c.Defaults.ImageProvider
}

func (c *Config) DefaultModel(providerID, capability string) string {
	p := c.Provider(providerID)
	if p == nil {
		return ""
	}
	capability = strings.ToLower(strings.TrimSpace(capability))
	if capability == "video" && p.Video != nil {
		return strings.TrimSpace(p.Video.DefaultModel)
	}
	if capability == "audio" && p.Audio != nil {
		return strings.TrimSpace(p.Audio.DefaultModel)
	}
	if capability == "image" && p.Image != nil {
		return strings.TrimSpace(p.Image.DefaultModel)
	}
	return ""
}

func (c *Config) Model(providerID, capability, modelID string) *ModelSpec {
	p := c.Provider(providerID)
	if p == nil {
		return nil
	}
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return nil
	}
	var spec *CapabilitySpec
	switch strings.ToLower(strings.TrimSpace(capability)) {
	case "video":
		spec = p.Video
	case "audio":
		spec = p.Audio
	default:
		spec = p.Image
	}
	if spec == nil {
		return nil
	}
	for i := range spec.Models {
		if strings.TrimSpace(spec.Models[i].ID) == modelID {
			return &spec.Models[i]
		}
	}
	return nil
}
