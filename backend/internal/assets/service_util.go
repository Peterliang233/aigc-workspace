package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func marshalParams(params any) string {
	if params == nil {
		return ""
	}
	b, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	return string(b)
}

func promptMeta(prompt string) (sha256Hex string, preview string) {
	prompt = strings.TrimSpace(prompt)
	sum := sha256.Sum256([]byte(prompt))
	sha256Hex = hex.EncodeToString(sum[:])

	preview = prompt
	if len([]rune(preview)) > 120 {
		preview = string([]rune(preview)[:120])
	}
	return sha256Hex, preview
}

