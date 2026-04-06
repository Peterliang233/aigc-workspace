package httpapi

import (
	"sort"
	"strconv"
	"strings"
)

func (h *Handler) animationDurationOptions(providerID, model string) []int {
	if h.models == nil {
		return []int{5}
	}
	ms := h.models.Model(providerID, "video", model)
	if ms == nil || ms.Form == nil {
		return []int{5}
	}
	seen := map[int]struct{}{}
	var out []int
	for _, f := range ms.Form.Fields {
		if !strings.EqualFold(strings.TrimSpace(f.Key), "duration_seconds") {
			continue
		}
		for _, opt := range f.Options {
			if n, err := strconv.Atoi(strings.TrimSpace(opt.Value)); err == nil && n > 0 {
				if _, ok := seen[n]; !ok {
					seen[n] = struct{}{}
					out = append(out, n)
				}
			}
		}
		if n, ok := asInt64(f.Default); ok && n > 0 {
			if _, exists := seen[int(n)]; !exists {
				seen[int(n)] = struct{}{}
				out = append(out, int(n))
			}
		}
	}
	if len(out) == 0 {
		return []int{5}
	}
	sort.Ints(out)
	return out
}
