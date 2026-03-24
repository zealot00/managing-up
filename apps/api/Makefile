DATABASE_URL ?= postgres://localhost:5432/skillhub?sslmode=disable

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: migrate
migrate: ## Run database migrations
	sqlx migrate run

.PHONY: migrate-down
migrate-down: ## Revert last database migration
	sqlx migrate revert

.PHONY: seed
seed: ## Seed the database
	go run cmd/seed/main.go

.PHONY: db-reset
db-reset: migrate-down migrate seed ## Reset database (revert, migrate, seed)

.PHONY: serve
serve: ## Run the server
	go run cmd/server/main.go

.PHONY: serve-pg
serve-pg: ## Run the server with DATABASE_URL
	DATABASE_URL=$(DATABASE_URL) go run cmd/server/main.go

.PHONY: build
build: ## Build the server binary
	go build -o bin/server ./cmd/server

.PHONY: test
test: ## Run all tests
	go test ./...
