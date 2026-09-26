.DEFAULT_GOAL := help
MIGRATE ?= migrate
GOFLAGS ?= -buildvcs=false
DATABASE_URL ?= postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSLMODE)

-include .env
export

.PHONY: help setup db-up db-down db-reset db-bootstrap migrate-up migrate-down migrate-down-docker seed api web test lint fmt
help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; printf "Scholarship OS commands:\n"} /^[a-zA-Z_-]+:.*?## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install backend and frontend dependencies
	cd apps/api && go mod download
	cd apps/web && npm install

db-up: ## Start PostgreSQL
	docker compose up -d postgres

db-down: ## Stop PostgreSQL
	docker compose down

db-reset: ## Delete local PostgreSQL data and restart (destructive)
	docker compose down -v
	docker compose up -d postgres

db-bootstrap: db-up ## Start PostgreSQL and apply migrations in Docker
	docker compose run --rm migrate

migrate-up: ## Apply all database migrations (requires golang-migrate)
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Roll back one database migration (requires golang-migrate)
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-down-docker: ## Roll back one migration using Docker
	docker compose run --rm migrate -path=/migrations -database="postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@postgres:5432/$(DATABASE_NAME)?sslmode=disable" down 1

seed: ## Load fictional Phase 3 catalog data into the Docker database
	docker compose exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -f /dev/stdin < db/seeds/phase3_development.sql

api: ## Run the Go API
	cd apps/api && go run ./cmd/server

web: ## Run the Vite frontend
	cd apps/web && npm run dev

test: ## Run backend tests and frontend checks
	cd apps/api && go test ./...
	cd apps/web && npm run build

lint: ## Run Go vet and frontend type checks
	cd apps/api && go vet ./...
	cd apps/web && npm run lint

fmt: ## Format backend and frontend sources
	cd apps/api && gofmt -w $$(find cmd internal pkg -type f -name '*.go')
	cd apps/web && npm run format
