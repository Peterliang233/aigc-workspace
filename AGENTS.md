# Repository Guidelines

## Project Structure

- `backend/`: Go API server (Gin) + MySQL persistence (GORM).
  - `backend/main.go`: server entry.
  - `backend/models.json`: provider/model list + per-model form requirements (no secrets).
  - `backend/internal/httpapi/`: HTTP routing, handlers, middleware.
  - `backend/internal/assets/`: generation history (MySQL) + asset ingestion (fetch vendor URL -> store in MinIO).
  - `backend/internal/blobstore/`: MinIO client wrapper.
  - `backend/internal/providers/`: vendor integrations (e.g. `siliconflow/`, `wuyinkeji/`, `openai_compatible/`).
- `frontend/`: React + Vite web app.
  - `frontend/src/components/`: `ImageStudio`, `VideoStudio`, `HistoryStudio`, `ToolboxStudio`.
- Root: `docker-compose*.yml`, `.env.example`, `Makefile`.

## Build, Test, and Development Commands

- MySQL (dev, Docker): `docker compose -f docker-compose.dev.yml up -d mysql`
  - Host port defaults to `3307` via `MYSQL_PORT` in root `.env`.
- MinIO (dev, Docker): `docker compose -f docker-compose.dev.yml up -d minio minio-init`
  - Console defaults to `http://localhost:9001` (see `MINIO_CONSOLE_PORT`).
- Backend (local): `cd backend && env GOCACHE=/tmp/gocache go run .`
- Frontend (local): `cd frontend && npm i && npm run dev`
- Tests:
  - Backend: `cd backend && go test ./...`
  - Frontend (typecheck + build): `cd frontend && npm run build`
- Compose (prod-like): `docker compose up -d --build`

## Coding Style & Naming

- Go: run `gofmt` on touched files (`gofmt -w ./backend/...`); keep packages under `backend/internal/...`.
- TypeScript/React: keep components in `frontend/src/components/`, prefer explicit names like `*Studio.tsx`.
- Logs/events: use consistent `slog` event names (e.g. `provider_*`, `downstream_*`).
- File size: keep any single `.go/.ts/.tsx/.css` file **<= 200 lines**. Split by concern (routes, services, helpers, UI subcomponents, stylesheets). Use `make check-lines`.

## Testing Guidelines

- Prefer small, table-driven Go tests close to the package under test.
- Any provider/handler change should still pass `go test ./...` and `npm run build`.

## Commit & Pull Request Guidelines

- Commits follow a simple Conventional Commits pattern seen in history: `feat: ...`, `fix: ...`.
- PRs should include:
  - What changed and why.
  - UI changes: screenshots (sidebar/results) when relevant.
  - Config changes: note required `.env` keys.

## Security & Configuration Notes

- Root `.env` is the single source for runtime config (do not maintain `backend/.env`).
- Any config key added to `.env.example` must be added to root `.env` in the same change, with a sensible default or empty placeholder.
- Do not `source .env` in a shell: DSNs like `tcp(127.0.0.1:3307)` are not POSIX shell syntax.
- Secrets (API keys) are env-managed; do not persist them in MySQL or log them.
- Downstream request logging records parameters; prompts are redacted by default. Only set `LOG_PROMPT_FULL=true` for local debugging.
- Logs are written to `log/` (gitignored) by default; override with `LOG_FILE`.
