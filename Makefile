.PHONY: help dev build run migrate-up migrate-down migrate-create sqlc lint test docker-up docker-down

# Default
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Run API with hot-reload (requires air: go install github.com/cosmtrek/air@latest)
	air -c .air.toml

build: ## Build the Go binary
	go build -o bin/api ./cmd/api

run: build ## Build and run
	./bin/api

# Database
migrate-up: ## Run migrations up
	migrate -path migrations -database "$(shell echo $${DATABASE_URL:-postgres://howallet:howallet_secret@localhost:5432/howallet?sslmode=disable})" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "$(shell echo $${DATABASE_URL:-postgres://howallet:howallet_secret@localhost:5432/howallet?sslmode=disable})" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=add_xyz)
	migrate create -ext sql -dir migrations -seq $(name)

# Code generation
sqlc: ## Generate sqlc code
	sqlc generate

# Quality
lint: ## Run linter
	golangci-lint run ./...

test: ## Run tests
	go test -v -race ./...

# Docker
docker-up: ## Start all services
	docker compose up -d --build

docker-down: ## Stop all services
	docker compose down

docker-logs: ## Tail logs
	docker compose logs -f
