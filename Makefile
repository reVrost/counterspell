.PHONY: build run dev clean test migrate generate install-deps install-air help

# Variables
BINARY_NAME=microscope-server
EXAMPLE_BINARY=microscope-example
GO_FILES=$(shell find . -name "*.go" -type f)
DB_PATH=./microscope.db

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Dependencies
install-deps: ## Install Go dependencies
	go mod download
	go mod tidy

install-air: ## Install air for hot reloading
	go install github.com/cosmtrek/air@latest

# Code generation
generate: ## Generate sqlc code
	sqlc generate

# Database operations
migrate: ## Run database migrations
	cd db && goose sqlite3 ../$(DB_PATH) up

migrate-down: ## Rollback last migration
	cd db && goose sqlite3 ../$(DB_PATH) down

migrate-status: ## Show migration status
	cd db && goose sqlite3 ../$(DB_PATH) status

migrate-reset: ## Reset database (drop all tables)
	cd db && goose sqlite3 ../$(DB_PATH) reset

# Build targets
build: generate ## Build the main server binary
	go build -o bin/$(BINARY_NAME) ./cmd/server

build-example: generate ## Build the example binary
	go build -o bin/$(EXAMPLE_BINARY) ./examples/basic

build-all: build build-example ## Build all binaries

# Run targets
run: build migrate ## Build and run the main server
	MICROSCOPE_AUTH_TOKEN=dev-token ./bin/$(BINARY_NAME)

run-example: build-example ## Build and run the example
	MICROSCOPE_AUTH_TOKEN=dev-token ./bin/$(EXAMPLE_BINARY)

# Development targets
dev: generate migrate ## Run server with hot reloading using air
	MICROSCOPE_AUTH_TOKEN=dev-token air

dev-example: generate migrate ## Run example with hot reloading
	air -c .air-example.toml

# Testing
test: generate ## Run tests
	go test -v ./...

test-race: generate ## Run tests with race detection
	go test -race -v ./...

test-coverage: generate ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Benchmarks
bench: generate ## Run benchmarks
	go test -bench=. -benchmem ./...

# Linting and formatting
fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run formatting and vetting

# Cleanup
clean: ## Clean build artifacts and database
	rm -rf bin/
	rm -f $(DB_PATH)
	rm -f coverage.out coverage.html

clean-db: ## Clean only the database
	rm -f $(DB_PATH)

# Docker targets (if needed later)
docker-build: ## Build Docker image
	docker build -t microscope:latest .

# Development utilities
logs: ## Tail server logs (if running in background)
	tail -f microscope.log

curl-health: ## Test health endpoint
	curl -H "Authorization: Bearer dev-token" http://localhost:1323/microscope/health

curl-logs: ## Test logs endpoint
	curl -H "Authorization: Bearer dev-token" "http://localhost:1323/microscope/api/logs?limit=10"

curl-traces: ## Test traces endpoint
	curl -H "Authorization: Bearer dev-token" "http://localhost:1323/microscope/api/traces?limit=10"

# Quick development setup
setup: install-deps generate migrate ## Full development setup
	@echo "✅ MicroScope development environment ready!"
	@echo "Run 'make dev' to start development server with hot reloading"
	@echo "Run 'make run' to start production server" 