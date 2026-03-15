package assets

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func makeObjectKey(capability, ext string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	date := time.Now().Format("2006/01/02")
	name := hex.EncodeToString(b[:])
	ext = strings.TrimSpace(ext)
	if ext == "" || strings.Contains(ext, "/") || strings.Contains(ext, "\\") {
		ext = ".bin"
	}
	// Keep keys path-like for easier lifecycle management.
	return filepath.ToSlash(fmt.Sprintf("%s/%s/%s%s", capability, date, name, ext)), nil
}

