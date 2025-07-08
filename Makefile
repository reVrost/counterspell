.PHONY: dev run test clean generate migrate help

# Default target
help: ## Show available commands
	@echo "Counterspell - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

# Development
dev: generate migrate ## Start development server with hot reloading
	@echo "🚀 Starting Counterspell in development mode..."
	COUNTERSPELL_AUTH_TOKEN=dev-token air

run: generate migrate ## Run the server
	@echo "🚀 Starting Counterspell server..."
	COUNTERSPELL_AUTH_TOKEN=dev-token go run ./cmd/server

# Code generation
generate: ## Generate sqlc code
	sqlc generate

# Database
migrate: ## Run database migrations
	@mkdir -p bin
	cd db && goose duckdb ../counterspell.db up

# Testing
test: generate ## Run all tests
	go test -v ./...

# Cleanup
clean: ## Clean build artifacts and database
	rm -rf bin/ counterspell.db coverage.out *.log 
