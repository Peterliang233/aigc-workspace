.PHONY: help dev dev-backend dev-frontend \
	backend-test backend-fmt backend-run \
	frontend-install frontend-dev frontend-build frontend-preview \
	env-init \
	mysql-up mysql-down mysql-logs mysql-reset \
	minio-up minio-down minio-logs minio-reset \
	compose-up compose-down compose-logs \
	compose-dev-up compose-dev-down compose-dev-logs

GOCACHE ?= /tmp/gocache
ENV_FILE ?= .env
ENV_EXAMPLE ?= .env.example

help:
	@printf "Targets:\\n"
	@printf "  dev                 Run backend + frontend\\n"
	@printf "  dev-backend         Run Go backend (GOCACHE=%s)\\n" "$(GOCACHE)"
	@printf "  dev-frontend        Run React frontend (Vite)\\n"
	@printf "  backend-test        Go tests\\n"
	@printf "  backend-fmt         gofmt\\n"
	@printf "  frontend-install    npm install\\n"
	@printf "  frontend-build      build frontend\\n"
	@printf "  mysql-up            docker compose (dev) up -d mysql\\n"
	@printf "  mysql-logs          docker compose (dev) logs mysql\\n"
	@printf "  mysql-down          docker compose (dev) stop mysql\\n"
	@printf "  mysql-reset         docker compose (dev) down -v (DELETES dev data)\\n"
	@printf "  minio-up            docker compose (dev) up -d minio (+ init bucket)\\n"
	@printf "  minio-logs          docker compose (dev) logs minio\\n"
	@printf "  minio-down          docker compose (dev) stop minio\\n"
	@printf "  minio-reset         docker compose (dev) down -v (DELETES dev data)\\n"
	@printf "  compose-up          docker compose (prod) up -d\\n"
	@printf "  compose-dev-up      docker compose (dev) up -d\\n"

dev:
	@$(MAKE) -j2 dev-backend dev-frontend

env-init:
	@test -f "$(ENV_FILE)" || (cp "$(ENV_EXAMPLE)" "$(ENV_FILE)" && printf "created %s from %s\\n" "$(ENV_FILE)" "$(ENV_EXAMPLE)")

dev-backend:
	@mkdir -p "$(GOCACHE)"
	# Do NOT `source` .env here: DSNs like tcp(127.0.0.1:3307) are not valid POSIX shell syntax.
	# The backend loads .env itself (walks upwards from CWD).
	cd backend && env GOCACHE="$(GOCACHE)" go run .

dev-frontend:
	cd frontend && npm run dev

backend-run: dev-backend

backend-test:
	@mkdir -p "$(GOCACHE)"
	cd backend && env GOCACHE="$(GOCACHE)" go test ./...

backend-fmt:
	cd backend && gofmt -w ./internal ./main.go

frontend-install:
	cd frontend && npm install

frontend-dev: dev-frontend

frontend-build:
	cd frontend && npm run build

frontend-preview:
	cd frontend && npm run preview -- --host 0.0.0.0 --port 4173

mysql-up:
	@$(MAKE) env-init
	docker compose -f docker-compose.dev.yml up -d mysql

mysql-logs:
	docker compose -f docker-compose.dev.yml logs --tail=200 -f mysql

mysql-down:
	docker compose -f docker-compose.dev.yml stop mysql

mysql-reset:
	# WARNING: this deletes the dev mysql volume (data) and re-initializes it.
	@$(MAKE) env-init
	docker compose -f docker-compose.dev.yml down -v

minio-up:
	@$(MAKE) env-init
	docker compose -f docker-compose.dev.yml up -d minio minio-init

minio-logs:
	docker compose -f docker-compose.dev.yml logs --tail=200 -f minio

minio-down:
	docker compose -f docker-compose.dev.yml stop minio

minio-reset:
	# WARNING: this deletes the dev minio volume (data) and re-initializes it.
	@$(MAKE) env-init
	docker compose -f docker-compose.dev.yml down -v

compose-up:
	@$(MAKE) env-init
	docker compose up -d --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs --tail=200 -f

compose-dev-up:
	@$(MAKE) env-init
	docker compose -f docker-compose.dev.yml up -d --build

compose-dev-down:
	docker compose -f docker-compose.dev.yml down

compose-dev-logs:
	docker compose -f docker-compose.dev.yml logs --tail=200 -f
