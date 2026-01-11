.PHONY: dev build test clean tidy generate lint docker-build docker-run

# Variables
PROJECT_NAME := counterspell
TARGET_MAIN := ./cmd/app
BINARY_DIR := ./local
BINARY_PATH := $(BINARY_DIR)/$(PROJECT_NAME)

# Default target
all: dev

##@ Development

dev: generate build
	@echo "Starting $(PROJECT_NAME) in development mode..."
	@direnv exec $(BINARY_PATH) -addr :8710 -db ./data/$(PROJECT_NAME).db

air:
	@echo "Starting air with direnv..."
	@./air-dev.sh

run: build
	@echo "Starting $(PROJECT_NAME)..."
	@direnv exec $(BINARY_PATH) -addr :8710 -db ./data/$(PROJECT_NAME).db

run: build
	@$(BINARY_PATH) -addr :8710 -db ./data/$(PROJECT_NAME).db

##@ Build

build: tidy
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_PATH) $(TARGET_MAIN)
	@echo "Binary built: $(BINARY_PATH)"

build-prod:
	@echo "Building $(PROJECT_NAME) for production..."
	@mkdir -p deploy
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o deploy/$(PROJECT_NAME) $(TARGET_MAIN)
	@echo "Production binary built: deploy/$(PROJECT_NAME)"

##@ Testing

test:
	@echo "Running tests..."
	@go test -v -race ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Note: golangci-lint not installed, skipping..."

##@ Dependencies

tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@go mod verify

deps:
	@echo "Installing dependencies..."
	@go install github.com/air-verse/air@latest
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

##@ Code Generation

generate: tidy
	@echo "Generating code..."
	@if command -v templ >/dev/null 2>&1; then \
		templ generate; \
	else \
		echo "templ not installed, skipping... (run 'make deps' to install)"; \
	fi

##@ Docker

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(PROJECT_NAME):latest .

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 3000:3000 -v $(PWD)/data:/app/data $(PROJECT_NAME):latest

##@ Cleanup

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@rm -rf deploy
	@rm -f coverage.out coverage.html

clean-all: clean
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
