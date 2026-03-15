package mock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigc-backend/internal/types"
)

type Provider struct {
	staticRoot string
}

func New(staticRoot string) *Provider {
	return &Provider{staticRoot: staticRoot}
}

func (p *Provider) ProviderName() string { return "mock" }

func (p *Provider) GenerateImage(ctx context.Context, req types.ImageGenerateRequest) (types.ImageGenerateResponse, error) {
	_ = ctx
	id := randHex(8)
	slog.Default().Debug("provider_mock_generate", "id", id)
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = "(empty prompt)"
	}

	// Simple SVG placeholder so frontend can render without external deps/API keys.
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%%" stop-color="#111827"/>
      <stop offset="100%%" stop-color="#0ea5e9"/>
    </linearGradient>
  </defs>
  <rect width="1024" height="1024" fill="url(#g)"/>
  <rect x="64" y="64" width="896" height="896" rx="36" fill="rgba(255,255,255,0.10)" stroke="rgba(255,255,255,0.25)"/>
  <text x="96" y="160" font-family="ui-sans-serif, system-ui, -apple-system" font-size="44" fill="white" opacity="0.92">Mock Image</text>
  <text x="96" y="230" font-family="ui-sans-serif, system-ui, -apple-system" font-size="26" fill="white" opacity="0.85">Prompt:</text>
  <foreignObject x="96" y="250" width="832" height="640">
    <div xmlns="http://www.w3.org/1999/xhtml" style="color:white; font-family: ui-sans-serif, system-ui, -apple-system; font-size: 22px; line-height: 1.4; opacity: 0.9; white-space: pre-wrap;">
      %s
    </div>
  </foreignObject>
  <text x="96" y="940" font-family="ui-sans-serif, system-ui, -apple-system" font-size="18" fill="white" opacity="0.7">%s</text>
</svg>
`, escapeXML(prompt), time.Now().Format(time.RFC3339))

	outDir := filepath.Join(p.staticRoot, "generated")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return types.ImageGenerateResponse{}, err
	}
	name := "img_" + id + ".svg"
	if err := os.WriteFile(filepath.Join(outDir, name), []byte(svg), 0o644); err != nil {
		return types.ImageGenerateResponse{}, err
	}

	return types.ImageGenerateResponse{
		ImageURLs: []string{"/static/generated/" + name},
		Provider:  p.ProviderName(),
	}, nil
}

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Avoid depending on the stdlib html package (some environments ship a trimmed GOROOT).
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
