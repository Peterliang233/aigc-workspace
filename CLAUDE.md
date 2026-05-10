# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AIGC Workspace is a web application that wraps multiple AIGC (AI-Generated Content) platform APIs—image, video, audio, and text generation—into a unified workbench. Go backend + React frontend, with MySQL for persistence and MinIO for asset storage.

## Development Commands

```bash
# Full dev stack (recommended)
make mysql-up && make minio-up && make -j2 dev

# Backend only
cd backend && env GOCACHE=/tmp/gocache go run .

# Frontend only
cd frontend && npm i && npm run dev

# Tests
cd backend && go test ./...          # Go tests
cd frontend && npm run build         # TypeScript check + Vite build

# Formatting & linting
gofmt -w ./backend/internal ./backend/main.go
make check-lines                     # Enforces <=200 lines per .go/.ts/.tsx/.css file

# Docker
make compose-dev-up                  # Dev mode (hot reload)
make compose-up                      # Production mode
```

## Architecture

**Backend** (Go 1.23, Gin, GORM): `backend/internal/`
- `httpapi/` — HTTP handlers and routing (`router.go` defines all endpoints)
- `providers/` — Vendor integrations (siliconflow, wuyinkeji/速创, gptbest/柏拉图)
- `assets/` — Generation history (MySQL) + asset archival (MinIO)
- `blobstore/` — MinIO client wrapper
- `modelcfg/` — Loads `backend/models.json` (provider/model/form config)
- `config/` — Env-based configuration from root `.env`
- `storyvideo/` — Story video pipeline (segmentation, planning, stitching)
- `mediaworker/` — Media worker HTTP server (ffprobe, slideshow generation)
- `textgen/` — Text generation types

**Frontend** (React 18, TypeScript, Vite): `frontend/src/`
- `components/` — Studio components per feature (ImageStudio, VideoStudio, AudioStudio, TextStudio, StoryVideoStudio, HistoryStudio)
- `state/` — Context API providers (generation.tsx, storyvideo.tsx)
- `hooks/` — Metadata fetchers (useImageMeta, useVideoMeta, etc.)
- `layout/` — Sidebar, tabs, icons

**Communication:** Frontend Vite dev server proxies `/api` → `localhost:8080`. Backend serves REST endpoints; videos use async polling (`POST /api/videos/jobs` → `GET /api/videos/jobs/{id}`).

## Key Conventions

- **File size limit:** Every `.go`, `.ts`, `.tsx`, `.css` file must be ≤200 lines. Split by concern.
- **Go packages:** All under `backend/internal/...`. Run `gofmt` on touched files.
- **Frontend components:** Named `*Studio.tsx` for top-level tabs. Sub-components in feature directories.
- **Logging:** Use `slog` with consistent event names (`provider_*`, `downstream_*`). Prompts redacted by default.
- **Configuration:** Root `.env` is the single source for runtime config. Model/form definitions live in `backend/models.json`. Never `source .env` in shell (DSNs contain non-POSIX syntax).
- **Commits:** Conventional Commits style (`feat:`, `fix:`).

## Configuration

- Provider API keys: env vars (`BLTCY_API_KEY`, `SILICONFLOW_API_KEY`, `WUYIN_API_KEY`)
- Model list + form requirements: `backend/models.json` (loaded at startup, served via `GET /api/meta/*`)
- MySQL: port 3307 on host (avoids conflict with local 3306)
- MinIO: API on 9000, console on 9001 (default creds: minioadmin/minioadmin)
- Media worker: port 8090

## Testing

- Go: table-driven tests close to the package under test. Run `go test ./...`.
- Frontend: `npm run build` serves as typecheck + build validation. No separate test runner.
- Any provider/handler change should pass both `go test ./...` and `npm run build`.
