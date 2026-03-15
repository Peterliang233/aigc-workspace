.PHONY: help dev dev-backend dev-frontend \
	backend-test backend-fmt backend-run \
	frontend-install frontend-dev frontend-build frontend-preview \
	mysql-up mysql-down mysql-logs \
	compose-up compose-down compose-logs \
	compose-dev-up compose-dev-down compose-dev-logs

GOCACHE ?= /tmp/gocache

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
	@printf "  compose-up          docker compose (prod) up -d\\n"
	@printf "  compose-dev-up      docker compose (dev) up -d\\n"

dev:
	@$(MAKE) -j2 dev-backend dev-frontend

dev-backend:
	@mkdir -p "$(GOCACHE)"
	cd backend && env GOCACHE="$(GOCACHE)" go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

backend-run: dev-backend

backend-test:
	@mkdir -p "$(GOCACHE)"
	cd backend && env GOCACHE="$(GOCACHE)" go test ./...

backend-fmt:
	cd backend && gofmt -w ./cmd ./internal

frontend-install:
	cd frontend && npm install

frontend-dev: dev-frontend

frontend-build:
	cd frontend && npm run build

frontend-preview:
	cd frontend && npm run preview -- --host 0.0.0.0 --port 4173

mysql-up:
	docker compose -f docker-compose.dev.yml up -d mysql

mysql-logs:
	docker compose -f docker-compose.dev.yml logs --tail=200 -f mysql

mysql-down:
	docker compose -f docker-compose.dev.yml stop mysql

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs --tail=200 -f

compose-dev-up:
	docker compose -f docker-compose.dev.yml up -d --build

compose-dev-down:
	docker compose -f docker-compose.dev.yml down

compose-dev-logs:
	docker compose -f docker-compose.dev.yml logs --tail=200 -f
