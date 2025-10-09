# Counterspell

A lightweight, embedded observability tool for Go applications that provides OpenTelemetry tracing and logging capabilities with a local SQLite database backend and REST API for data querying.

**⚠️ This project is a work in progress and is not yet ready for production use. ⚠️**

## What it is

- Fast and easy to get started, no added extra cost
- Gives you observability UI for your LLM calls with the greatest of ease
- Prompt evals
- Prompt optimizer
- Embedded observability with otel, zerolog (uses sqlite), throwaway sqlite db
- Means no external dependencies, no xtra docker containers
- Writes logs on a separate goroutine, so your app is not affected

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go (1.25 or later)
- Node.js (20.x or later)

### Running with Docker Compose

The easiest way to get Counterspell up and running is with Docker Compose.

```bash
docker-compose up
```

This will start the backend server on port `8080` and the frontend UI on port `5173`.

- **Backend API**: `http://localhost:8080`
- **Frontend UI**: `http://localhost:5173`

### Development Environment

For development, you can use the provided `dev.sh` script, which uses `kitty` to create a split-pane terminal for the backend and frontend.

```bash
./dev.sh
```

# Counterspell

## Installation

```bash
go get github.com/revrost/counterspell
```

## Todo

- [ ] Protobuf schemas
- [ ] Agent configuration/blueprint framework
- [ ] Openrouter integration
- [ ] Lightweight execution runtime utilize goroutine (cadence/go-workflow)
- [ ] Create agent, run agent, watch UI
- [ ] Orchestrator-Executor MVP via ui
- [ ] Otel integration
- [ ] Openapi streaming spec
- [ ] Move sqlite to postgres (later)

## Quick Start

The simplest way to add Counterspell to your application:

```go
func main() {
	// Example 1: Using Echo router
	err := counterspell.AddToEcho(e,
		counterspell.WithAuthToken("secret"),
	)
	if err != nil {
		slog.Error("Error adding counterspell middleware", "err", err)
	}
}

Go to your app endpoint e.g localhost:8080/counterspell
```

## API Endpoints

All API endpoints require authentication via the `Authorization: Bearer <token>` header or `auth` query parameter.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the [Apache-2.0 License](LICENSE).
