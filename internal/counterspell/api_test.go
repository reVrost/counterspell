package counterspell

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/revrost/counterspell/internal/db"
)

func setupAPITestDB(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	_, err = database.Exec(`
		CREATE TABLE logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			trace_id TEXT,
			span_id TEXT,
			attributes TEXT
		);
		CREATE INDEX idx_logs_timestamp ON logs(timestamp);
		CREATE INDEX idx_logs_level ON logs(level);
		CREATE INDEX idx_logs_trace_id ON logs(trace_id);

		CREATE TABLE spans (
			span_id TEXT PRIMARY KEY,
			trace_id TEXT NOT NULL,
			parent_span_id TEXT,
			name TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			duration_ns INTEGER NOT NULL,
			attributes TEXT,
			service_name TEXT NOT NULL,
			has_error BOOLEAN NOT NULL DEFAULT FALSE
		);
		CREATE INDEX idx_spans_trace_id ON spans(trace_id);
		CREATE INDEX idx_spans_start_time ON spans(start_time);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return database
}

func insertTestLogs(t *testing.T, database *sql.DB) {
	queries := db.New(database)

	// Insert test logs
	logs := []db.InsertLogParams{
		{
			Timestamp:  "2024-01-15T10:30:00.123Z",
			Level:      "info",
			Message:    "Test info log",
			TraceID:    sql.NullString{String: "trace123", Valid: true},
			SpanID:     sql.NullString{String: "span123", Valid: true},
			Attributes: sql.NullString{String: `{"user":"test"}`, Valid: true},
		},
		{
			Timestamp:  "2024-01-15T10:31:00.123Z",
			Level:      "error",
			Message:    "Test error log",
			TraceID:    sql.NullString{String: "trace456", Valid: true},
			SpanID:     sql.NullString{String: "span456", Valid: true},
			Attributes: sql.NullString{String: `{"error":"test error"}`, Valid: true},
		},
		{
			Timestamp:  "2024-01-15T10:32:00.123Z",
			Level:      "debug",
			Message:    "Test debug log",
			Attributes: sql.NullString{String: `{"debug":"test"}`, Valid: true},
		},
	}

	for _, log := range logs {
		if err := queries.InsertLog(context.Background(), log); err != nil {
			t.Fatalf("Failed to insert test log: %v", err)
		}
	}
}

func insertTestSpans(t *testing.T, database *sql.DB) {
	queries := db.New(database)

	// Insert test spans
	spans := []db.InsertSpanParams{
		{
			SpanID:      "span123",
			TraceID:     "trace123",
			Name:        "GET /hello",
			StartTime:   "2024-01-15T10:30:00.000Z",
			EndTime:     "2024-01-15T10:30:00.100Z",
			DurationNs:  100000000,
			Attributes:  sql.NullString{String: `{"http.method":"GET"}`, Valid: true},
			ServiceName: "test-service",
			HasError:    false,
		},
		{
			SpanID:      "span456",
			TraceID:     "trace456",
			Name:        "POST /error",
			StartTime:   "2024-01-15T10:31:00.000Z",
			EndTime:     "2024-01-15T10:31:00.200Z",
			DurationNs:  200000000,
			Attributes:  sql.NullString{String: `{"http.method":"POST"}`, Valid: true},
			ServiceName: "test-service",
			HasError:    true,
		},
		{
			SpanID:      "span789",
			TraceID:     "trace789",
			Name:        "Background Task",
			StartTime:   "2024-01-15T10:32:00.000Z",
			EndTime:     "2024-01-15T10:32:00.500Z",
			DurationNs:  500000000,
			Attributes:  sql.NullString{String: `{"task":"background"}`, Valid: true},
			ServiceName: "worker-service",
			HasError:    false,
		},
	}

	for _, span := range spans {
		if err := queries.InsertSpan(context.Background(), span); err != nil {
			t.Fatalf("Failed to insert test span: %v", err)
		}
	}
}

func createTestEcho(database *sql.DB, authToken string) *echo.Echo {
	e := echo.New()
	handler := NewAPIHandler(database)

	// Custom middleware to check for secret query parameter
	secretAuth := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			secret := c.QueryParam("secret")
			if secret == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "secret query parameter is required")
			}
			if secret != authToken {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid secret")
			}
			return next(c)
		}
	}

	api := e.Group("/counterspell/api", secretAuth)

	api.GET("/logs", handler.QueryLogs)
	api.GET("/traces", handler.QueryTraces)
	api.GET("/traces/:trace_id", handler.GetTraceDetails)

	// Add health endpoint (no auth required)
	e.GET("/counterspell/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "counterspell",
		})
	})

	return e
}

func TestAPIHandler_QueryLogs(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestLogs(t, database)

	e := createTestEcho(database, "test-token")

	// Test successful request with secret parameter
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?limit=10&secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Metadata["total"] != float64(3) {
		t.Errorf("Expected 3 total logs, got %v", response.Metadata["total"])
	}

	logs, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", response.Data)
	}
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs in data, got %d", len(logs))
	}
}

func TestAPIHandler_QueryLogsWithFilters(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestLogs(t, database)

	e := createTestEcho(database, "test-token")

	// Test level filter
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?level=error&secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	logs, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", response.Data)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(logs))
	}
}

func TestAPIHandler_QueryLogsUnauthorized(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	e := createTestEcho(database, "test-token")

	// Test request without secret parameter
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAPIHandler_QueryLogsWrongToken(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	e := createTestEcho(database, "test-token")

	// Test request with wrong secret
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?secret=wrong-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAPIHandler_QueryTraces(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestSpans(t, database)

	e := createTestEcho(database, "test-token")

	// Test successful request
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/traces?limit=10&secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	traces, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", response.Data)
	}

	// Should have 3 traces (root spans)
	if len(traces) != 3 {
		t.Errorf("Expected 3 traces, got %d", len(traces))
	}
}

func TestAPIHandler_QueryTracesWithErrorFilter(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestSpans(t, database)

	e := createTestEcho(database, "test-token")

	// Test has_error filter
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/traces?has_error=true&secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	traces, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", response.Data)
	}

	if len(traces) != 1 {
		t.Errorf("Expected 1 trace with error, got %d", len(traces))
	}
}

func TestAPIHandler_GetTraceDetails(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestSpans(t, database)

	e := createTestEcho(database, "test-token")

	// Test successful request
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/traces/trace123?secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response TraceDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.TraceID != "trace123" {
		t.Errorf("Expected trace_id 'trace123', got '%s'", response.TraceID)
	}

	if len(response.Spans) != 1 {
		t.Errorf("Expected 1 span for trace123, got %d", len(response.Spans))
	}

	span := response.Spans[0]
	if span.Name != "GET /hello" {
		t.Errorf("Expected span name 'GET /hello', got '%s'", span.Name)
	}
	if span.HasError {
		t.Errorf("Expected span to not have error, but it does")
	}
}

func TestAPIHandler_GetTraceDetailsNotFound(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	e := createTestEcho(database, "test-token")

	// Test non-existent trace
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/traces/nonexistent?secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestAPIHandler_HealthEndpoint(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	e := createTestEcho(database, "test-token")

	// Test health endpoint (no auth required)
	req := httptest.NewRequest(http.MethodGet, "/counterspell/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
	if response["service"] != "counterspell" {
		t.Errorf("Expected service 'counterspell', got '%s'", response["service"])
	}
}

func TestAPIHandler_QueryLogsTextSearch(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestLogs(t, database)

	e := createTestEcho(database, "test-token")

	// Test text search
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?q=error&secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	logs, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", response.Data)
	}

	// Should find logs with "error" in message or attributes
	if len(logs) == 0 {
		t.Error("Expected to find logs with 'error' in search")
	}
}

func TestAPIHandler_QueryLogsPagination(t *testing.T) {
	database := setupAPITestDB(t)
	defer database.Close()

	insertTestLogs(t, database)

	e := createTestEcho(database, "test-token")

	// Test pagination
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?limit=2&offset=1&secret=test-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Metadata["limit"] != float64(2) {
		t.Errorf("Expected limit 2, got %v", response.Metadata["limit"])
	}
	if response.Metadata["offset"] != float64(1) {
		t.Errorf("Expected offset 1, got %v", response.Metadata["offset"])
	}

	logs, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", response.Data)
	}

	if len(logs) > 2 {
		t.Errorf("Expected at most 2 logs due to limit, got %d", len(logs))
	}
}
