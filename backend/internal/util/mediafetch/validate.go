package mediafetch

import (
	"errors"
	"mime"
	"net"
	"net/url"
	"path/filepath"
	"strings"
)

func validateURL(rawURL string) error {
	pu, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if pu.Scheme != "https" && pu.Scheme != "http" {
		return errors.New("refusing to download non-http(s) url")
	}
	host := strings.ToLower(pu.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return errors.New("refusing to download from localhost")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
			return errors.New("refusing to download from local/private ip")
		}
	}
	return nil
}

func GuessExt(contentType, rawURL string) string {
	if contentType != "" {
		mt, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			switch strings.ToLower(mt) {
			case "image/png":
				return ".png"
			case "image/jpeg":
				return ".jpg"
			case "image/webp":
				return ".webp"
			case "image/gif":
				return ".gif"
			case "image/svg+xml":
				return ".svg"
			}
		}
	}

	pu, err := url.Parse(rawURL)
	if err == nil {
		ext := strings.ToLower(filepath.Ext(pu.Path))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg":
			if ext == ".jpeg" {
				return ".jpg"
			}
			return ext
		}
	}
	return ""
}

