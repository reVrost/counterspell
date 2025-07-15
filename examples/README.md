# Counterspell Examples

This directory contains examples demonstrating how to integrate Counterspell observability with different web frameworks. All examples use **DuckDB** as the storage backend for logs and traces.

## Examples Overview

### 1. Echo v4 (`/echo`)
Demonstrates integration with Echo v4 framework:
```bash
cd echo
go run main.go
```
- **Port**: 8080
- **Auth Token**: `my-secret-token`
- **Database**: `counterspell_echo.db`

### 2. Standard Library (`/stdlib`) 
Demonstrates integration with Go's standard library HTTP router:
```bash
cd stdlib
go run main.go
```
- **Port**: 8081
- **Auth Token**: `my-other-secret-token`  
- **Database**: `counterspell_stdlib.db`

### 3. Echo v5 (`/echo-v5`)
Demonstrates integration with Echo v5 framework:
```bash
cd echo-v5
go mod tidy
go run main.go
```
- **Port**: 8082
- **Auth Token**: `my-echo-v5-token`
- **Database**: `counterspell_echo_v5.db`

## Key Features Demonstrated

### 🔍 **Observability**
- **Structured Logging**: JSON logs with trace context
- **Distributed Tracing**: OpenTelemetry spans with correlation
- **Error Tracking**: Automatic error capture and logging

### 🗃️ **DuckDB Storage**
- **High Performance**: Fast analytical queries on logs/traces
- **Simple Deployment**: Single file database (no external dependencies)
- **Rich Queries**: SQL-based log and trace analysis

### 🌐 **Web UI & API**
- **Health Endpoint**: `/counterspell/health`
- **Log API**: `/counterspell/api/logs?secret=<token>`
- **Trace API**: `/counterspell/api/traces?secret=<token>`
- **Web UI**: `/counterspell/` (embedded React app)

## Testing the Examples

### Generate Some Data
```bash
# Hit some endpoints to generate logs and traces
curl http://localhost:8080/hello
curl http://localhost:8082/slow
```

### View Health Status
```bash
curl http://localhost:8080/counterspell/health
```

### Query Logs
```bash
curl "http://localhost:8080/counterspell/api/logs?secret=my-secret-token"
```

### Query Traces  
```bash
curl "http://localhost:8080/counterspell/api/traces?secret=my-secret-token"
```

## Migration from SQLite

These examples showcase the **migration from SQLite to DuckDB**:

- ✅ **Better Performance**: DuckDB excels at analytical workloads
- ✅ **Simplified Deployment**: No CGO dependencies like SQLite
- ✅ **Advanced Analytics**: Native columnar storage for fast queries
- ✅ **Same API**: Drop-in replacement with identical functionality

## Framework Support

Counterspell provides dedicated functions for each framework:

```go
// Echo v4
counterspell.AddToEcho(e, opts...)

// Echo v5  
counterspell.AddToEchoV5(e, opts...)

// Standard Library
counterspell.AddToStdlib(mux, opts...)
```

All functions support the same configuration options:
- `WithDBPath(path)` - DuckDB database file location
- `WithAuthToken(token)` - API authentication
- `WithServiceName(name)` - Service identification
- `WithServiceVersion(version)` - Version tracking

## Database Files

Each example creates its own DuckDB database file:
- `counterspell_echo.db` + `.wal` files
- `counterspell_stdlib.db` + `.wal` files  
- `counterspell_echo_v5.db` + `.wal` files

The `.wal` (Write-Ahead Log) files are normal DuckDB operation files for transaction safety. 