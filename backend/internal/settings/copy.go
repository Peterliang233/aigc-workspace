package settings

func deepCopy(s Settings) Settings {
	out := Settings{}
	if s.ImageProviders != nil {
		out.ImageProviders = map[string]ProviderSettings{}
		for k, v := range s.ImageProviders {
			out.ImageProviders[k] = copyProv(v)
		}
	}
	return out
}

func copyProv(p ProviderSettings) ProviderSettings {
	out := ProviderSettings{}
	if p.BaseURL != nil {
		v := *p.BaseURL
		out.BaseURL = &v
	}
	if p.APIKey != nil {
		v := *p.APIKey
		out.APIKey = &v
	}
	if p.DefaultModel != nil {
		v := *p.DefaultModel
		out.DefaultModel = &v
	}
	if p.Models != nil {
		cp := make([]string, 0, len(*p.Models))
		for _, m := range *p.Models {
			cp = append(cp, m)
		}
		out.Models = &cp
	}
	return out
}
