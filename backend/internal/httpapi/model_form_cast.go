package httpapi

import (
	"fmt"
	"strings"
)

func asString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		return s, s != ""
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		return s, s != ""
	}
}

func asInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, false
		}
		var n int64
		_, err := fmt.Sscan(s, &n)
		return n, err == nil
	default:
		var n int64
		_, err := fmt.Sscan(fmt.Sprint(v), &n)
		return n, err == nil
	}
}
