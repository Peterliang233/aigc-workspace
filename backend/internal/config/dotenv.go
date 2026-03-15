package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadDotEnv is a tiny .env loader to avoid pulling external deps.
// It only supports simple KEY=VALUE lines and ignores comments/blank lines.
// Existing environment variables are not overwritten.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" {
			continue
		}
		if _, exists := os.LookupEnv(k); exists {
			continue
		}
		// Strip surrounding quotes when present.
		if len(v) >= 2 {
			if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
				v = v[1 : len(v)-1]
			}
		}
		_ = os.Setenv(k, v)
	}
}

// LoadDotEnvUpwards searches for ".env" starting from the current working
// directory and walking upwards, and loads the first one found.
//
// This lets developers run the backend from either repo root or backend/
// without maintaining multiple .env files.
func LoadDotEnvUpwards(maxHops int) {
	if maxHops <= 0 {
		maxHops = 8
	}
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	for i := 0; i < maxHops; i++ {
		p := filepath.Join(dir, ".env")
		if _, err := os.Stat(p); err == nil {
			LoadDotEnv(p)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}
