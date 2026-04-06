package wuyinkeji

import (
	"fmt"
	"strings"
)

func extraString(extra map[string]any, key string) string {
	if len(extra) == 0 {
		return ""
	}
	value, ok := extra[key]
	if !ok || value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" || strings.EqualFold(text, "null") {
		return ""
	}
	return text
}
