.PHONY: help setup dev build run clean test fmt lint vet migrate-up migrate-down migrate-create migrate-status sqlc generate db-drop env mock-gen mock-update ui-build ui-dev ui-clean build dev

# Application
APP_NAME=invoker
CMD_PATH=./cmd/invoker
ECR_REGISTRY := 938514907939.dkr.ecr.ap-southeast-2.amazonaws.com
IMAGE_NAME := invoker
FULL_IMAGE_NAME := $(ECR_REGISTRY)/$(IMAGE_NAME)
BUILD_PATH=./bin/$(APP_NAME)

# Database
SCHEMA_FILE=schema.sql
DB_URL?=${DATABASE_URL}

# Tools
SQLC?=sqlc
MOCKGEN?=go run go.uber.org/mock/mockgen@latest

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(APP_NAME) ./cmd/invoker/main.go
	@echo "Binary built: $(APP_NAME)"

dev: build
	ENV=dev ./$(APP_NAME)

setup: ## Set up development environment
	@echo "Setting up development environment..."
	@echo "1. Installing Go dependencies..."
	@go mod download
	@go mod verify
	@echo "2. Checking sqlc installation..."
	@which $(SQLC) > /dev/null || (echo "sqlc not found. Install from https://docs.sqlc.dev/en/stable/overview/install.html" && exit 1)
	@echo "3. Copying .env.example to .env..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "✓ Created .env file (please edit with your credentials)"; fi
	@echo "✓ Development environment setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit .env with your Supabase and database credentials"
	@echo "2. Run 'make migrate-up' to set up the database"
	@echo "3. Run 'make sqlc' to generate database code"
	@echo "4. Run 'make dev' to start API development server"
	@echo "5. Run 'make ui' to start UI development server"

env: ## Show current environment configuration
	@echo "Environment Configuration:"
	@echo "=========================="
	@if [ -f .env ]; then \
		echo "Using .env file"; \
		grep -v "^#" .env | grep -v "^$$"; \
	else \
		echo "No .env file found. Using system environment variables."; \
	fi

deps: ## Install/update dependencies
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

sqlc: ## Generate database code using sqlc
	@echo "Generating database code..."
	@which $(SQLC) > /dev/null || (echo "sqlc not found. Install from https://docs.sqlc.dev/en/stable/overview/install.html" && exit 1)
	@$(SQLC) generate
	@echo "✓ Database code generated"

mock-gen: ## Generate mocks for all interfaces
	@echo "Generating mocks..."
	@$(MOCKGEN) -source=internal/db/service.go -destination=internal/db/mock_service.go -package=db
	@echo "✓ Mocks generated"

mock-update: mock-gen ## Alias for mock-gen

migrate-up: ## Run database migrations using golang-migrate
	@echo "Running database migrations with golang-migrate..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DATABASE_URL not set"; \
		echo "Set it in .env or run: export DATABASE_URL=your-db-url"; \
		exit 1; \
	fi
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -database $(DB_URL) -path ./migrations up

migrate-down: ## Rollback last migration using golang-migrate
	@echo "Rolling back last migration..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DATABASE_URL not set"; \
		exit 1; \
	fi
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -database $(DB_URL) -path ./migrations down 1

migrate-create: ## Create a new migration file (usage: make migrate-create name=your_migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter required"; \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -database $(DB_URL) -path ./migrations create -ext sql $(name)

migrate-status: ## Show migration status using golang-migrate
	@echo "Checking migration status..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DATABASE_URL not set"; \
		exit 1; \
	fi
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -database $(DB_URL) -path ./migrations version

db-reset: ## Reset database (drop and recreate)
	@echo "WARNING: This will drop all database tables!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		psql $(DB_URL) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" && echo "✓ Database schema dropped"; \
	else \
		echo "Operation cancelled"; \
	fi
	@echo "Running migrations..."
	@$(MAKE) migrate-up


docker-build:
	docker build --platform linux/amd64 -t $(IMAGE_NAME) .

docker-tag:
	docker tag $(IMAGE_NAME):latest $(FULL_IMAGE_NAME):latest

docker-login:
	aws ecr get-login-password --region ap-southeast-2 | docker login --username AWS --password-stdin $(ECR_REGISTRY)

docker-push:
	docker push $(FULL_IMAGE_NAME):latest
