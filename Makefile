.PHONY: dev build test clean tidy generate lint docker-build docker-run run ui preprod kill-dev

# Variables
PROJECT_NAME := counterspell
TARGET_MAIN := ./cmd/app
BINARY_PATH := $(PROJECT_NAME)

# Default target
all: dev

##@ Development

# dev: generate build
# 	@echo "Starting $(PROJECT_NAME) in development mode..."
# 	./counterspell -addr :8710 -db ./data/$(PROJECT_NAME).db
#
kill-dev:
	kill -9 $$(lsof -t -i:8710) 2>/dev/null || true

dev: build
	ENV=dev ./$(PROJECT_NAME) -addr :8710

ui: ## Run Vite dev server (frontend on :5173, proxies to Go on :8710)
	@echo "Starting Vite dev server..."
	@FRONTEND_URL=http://localhost:5173 cd ui && npm run dev

preprod: build ## Run Go server with built frontend (everything on :8710)
	@echo "Building frontend..."
	@cd ui && npm run build
	@echo "Starting $(PROJECT_NAME) in preprod mode..."
	@FRONTEND_URL=http://localhost:8710 ./$(BINARY_PATH)

air:
	@echo "Starting air with direnv..."
	@./air-dev.sh

run: build
	@echo "Starting $(PROJECT_NAME)..."
	@direnv exec $(BINARY_PATH) -addr :8710 -db ./data/$(PROJECT_NAME).db

##@ Build

build:
	@echo "Building $(PROJECT_NAME)..."
	@go build -o $(BINARY_PATH) $(TARGET_MAIN)
	@echo "Binary built: $(BINARY_PATH)"

build-prod: tidy
	@echo "Building $(PROJECT_NAME) for production..."
	@mkdir -p deploy
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o deploy/$(PROJECT_NAME) $(TARGET_MAIN)
	@echo "Production binary built: deploy/$(PROJECT_NAME)"

##@ Testing

test:
	@echo "Running tests..."
	@go test -v ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-e2e:
	@echo "Running E2E tests..."
	@cd ui && npm run test:e2e

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Note: golangci-lint not installed, skipping..."

check-all: lint
	@echo "Running all checks..."
	@cd ui && npm run check

format:
	@echo "Formatting code..."
	@cd ui && npx prettier --write .
	@go fmt ./...

##@ Dependencies

tidy:
	@go mod tidy
	@go mod verify

deps:
	@echo "Installing dependencies..."
	@go install github.com/air-verse/air@latest
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

##@ Code Generation

generate: sqlc
	@echo "Generating code..."
	@if command -v templ >/dev/null 2>&1; then \
		templ generate; \
	else \
		echo "templ not installed, skipping... (run 'make deps' to install)"; \
	fi
	go generate ./...

sqlc:
	@echo "Generating sqlc code..."
	@if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	else \
		echo "sqlc not installed, skipping... (run 'go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest')"; \
	fi

##@ Docker

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(PROJECT_NAME):latest .

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 3000:3000 -v $(PWD)/data:/app/data $(PROJECT_NAME):latest

##@ Cleanup


clean-all:
	@echo "Cleaning all generated files..."
	@rm -rf data/*.db data/*.db-shm data/*.db-wal
	@rm -rf worktree-*

##@ Database

migrate-up:
	@echo "Running database migrations..."
	@sqlite3 data/$(PROJECT_NAME).db < internal/db/schema.sql

migrate-down:
	@echo "Dropping database..."
	@rm -f data/$(PROJECT_NAME).db

##@ Help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
