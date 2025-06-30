# MicroScope

A lightweight, embedded observability tool for Go applications that provides OpenTelemetry tracing and logging capabilities with a local SQLite database backend and REST API for data querying.

## Features

- **Self-contained**: No external dependencies or services required
- **OpenTelemetry Integration**: Automatic instrumentation for tracing
- **Custom Logging**: Seamless integration with zerolog for structured logging
- **SQLite Backend**: Local database storage for traces and logs
- **REST API**: Query traces and logs via HTTP endpoints
- **One-liner Integration**: Simple installation with functional options
- **Graceful Shutdown**: Automatic cleanup when your application stops

## Installation

```bash
go get github.com/your-github-username/microscope
```

## Quick Start

The simplest way to add MicroScope to your application:

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/rs/zerolog/log"
    "github.com/your-github-username/microscope"
    "go.opentelemetry.io/otel"
)

func main() {
    e := echo.New()

    // One-liner installation - set MICROSCOPE_AUTH_TOKEN environment variable
    if err := microscope.Install(e); err != nil {
        log.Fatal().Err(err).Msg("Failed to install MicroScope")
    }

    // Your application routes
    tracer := otel.Tracer("my-app")
    e.GET("/hello", func(c echo.Context) error {
        _, span := tracer.Start(c.Request().Context(), "hello-handler")
        defer span.End()

        log.Info().Str("user", "demo").Msg("Request received")
        return c.String(200, "Hello, World!")
    })

    log.Info().Msg("Server starting on :1323")
    log.Info().Msg("MicroScope API available at /microscope/api")
    e.Logger.Fatal(e.Start(":1323"))
}
```

## Configuration

MicroScope supports configuration through functional options:

```go
err := microscope.Install(e,
    microscope.WithAuthToken("my-secret-token"),
    microscope.WithDBPath("./data/observability.db"),
)
```

### Environment Variables

- `MICROSCOPE_AUTH_TOKEN`: Authentication token for API access (required)

## API Endpoints

All API endpoints require authentication via the `Authorization: Bearer <token>` header or `auth` query parameter.

### Health Check

```
GET /microscope/health
```

Returns the health status of MicroScope (no authentication required).

### Query Logs

```
GET /microscope/api/logs?limit=100&offset=0&level=info&q=search&trace_id=...
```

**Query Parameters:**
- `limit`: Number of logs to return (default: 100)
- `offset`: Offset for pagination (default: 0)
- `level`: Filter by log level (debug, info, warn, error)
- `q`: Full-text search on message and attributes
- `start_time`: Filter logs after this timestamp (RFC3339 format)
- `end_time`: Filter logs before this timestamp (RFC3339 format)
- `trace_id`: Filter logs by trace ID

**Response:**
```json
{
  "metadata": {
    "total": 1250,
    "limit": 100,
    "offset": 0
  },
  "data": [
    {
      "id": 1,
      "timestamp": "2024-01-15T10:30:00Z",
      "level": "info",
      "message": "Request received",
      "trace_id": "abc123...",
      "span_id": "def456...",
      "attributes": {
        "user": "demo",
        "endpoint": "/hello"
      }
    }
  ]
}
```

### Query Traces

```
GET /microscope/api/traces?limit=50&offset=0&q=search&has_error=true
```

**Query Parameters:**
- `limit`: Number of traces to return (default: 100)
- `offset`: Offset for pagination (default: 0)
- `q`: Search in root span names
- `has_error`: Filter traces that have errors (true/false)

**Response:**
```json
{
  "metadata": {
    "total": 45,
    "limit": 50,
    "offset": 0
  },
  "data": [
    {
      "trace_id": "abc123...",
      "root_span_name": "GET /hello",
      "trace_start_time": "2024-01-15T10:30:00Z",
      "duration_ms": 125.5,
      "span_count": 3,
      "error_count": 0,
      "has_error": false
    }
  ]
}
```

### Get Trace Details

```
GET /microscope/api/traces/{trace_id}
```

**Response:**
```json
{
  "trace_id": "abc123...",
  "spans": [
    {
      "span_id": "def456...",
      "trace_id": "abc123...",
      "parent_span_id": null,
      "name": "GET /hello",
      "start_time": "2024-01-15T10:30:00.000Z",
      "end_time": "2024-01-15T10:30:00.125Z",
      "duration_ns": 125500000,
      "attributes": {
        "http.method": "GET",
        "http.url": "/hello"
      },
      "service_name": "my-app",
      "has_error": false
    }
  ]
}
```

## Database Schema

MicroScope uses SQLite with the following schema:

### Logs Table

```sql
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    trace_id TEXT,
    span_id TEXT,
    attributes TEXT -- JSON string
);
```

### Spans Table

```sql
CREATE TABLE spans (
    span_id TEXT PRIMARY KEY,
    trace_id TEXT NOT NULL,
    parent_span_id TEXT,
    name TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    duration_ns INTEGER NOT NULL,
    attributes TEXT, -- JSON string
    service_name TEXT NOT NULL,
    has_error BOOLEAN NOT NULL DEFAULT FALSE
);
```

## Development

### Prerequisites

- Go 1.21 or later
- SQLite3

### Building

```bash
# Build the example server
go build ./cmd/server

# Build the basic example
go build ./examples/basic

# Run tests
go test ./...
```

### Database Migrations

Migrations are handled automatically using Goose when MicroScope starts. The migration files are embedded in the binary.

### Code Generation

Database access code is generated using `sqlc`:

```bash
# Generate database code (if you modify queries)
sqlc generate
```

## Architecture

MicroScope consists of several key components:

1. **Custom OpenTelemetry Exporter**: Writes spans to SQLite asynchronously
2. **Custom Zerolog Writer**: Writes logs to SQLite with trace correlation
3. **Database Layer**: Uses sqlc for type-safe database operations
4. **API Layer**: REST endpoints for querying observability data
5. **Migrations**: Automatic database schema management with Goose

The system is designed to be:
- **Non-blocking**: All database writes happen asynchronously
- **Performant**: Batched writes to minimize database overhead
- **Reliable**: Graceful shutdown ensures data integrity

## License

Apache License 2.0

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

- GitHub Issues: Report bugs and request features
- Documentation: See the [docs](./docs) directory for detailed guides
- Examples: Check the [examples](./examples) directory for more usage patterns
