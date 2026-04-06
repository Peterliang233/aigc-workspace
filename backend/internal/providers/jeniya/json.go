package jeniya

import (
	"encoding/json"
	"fmt"
	"strings"
)

func findString(raw []byte, keys ...string) string {
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return ""
	}
	return walkString(data, keys...)
}

func walkString(v any, keys ...string) string {
	switch x := v.(type) {
	case map[string]any:
		for _, key := range keys {
			for mk, mv := range x {
				if strings.EqualFold(strings.TrimSpace(mk), strings.TrimSpace(key)) {
					if s := scalarString(mv); s != "" {
						return s
					}
				}
			}
		}
		for _, mv := range x {
			if s := walkString(mv, keys...); s != "" {
				return s
			}
		}
	case []any:
		for _, mv := range x {
			if s := walkString(mv, keys...); s != "" {
				return s
			}
		}
	}
	return ""
}

func scalarString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case bool, float64, float32, int, int64, uint64:
		return strings.TrimSpace(fmt.Sprint(x))
	default:
		return ""
	}
}

func normalizeStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "succeeded", "completed", "complete", "done":
		return "succeeded"
	case "failed", "error", "cancelled", "canceled":
		return "failed"
	case "pending", "queued", "created":
		return "queued"
	default:
		return "running"
	}
}
