package httpapi

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func removeStr(list []string, v string) []string {
	var out []string
	for _, x := range list {
		if x == v {
			continue
		}
		out = append(out, x)
	}
	return out
}
