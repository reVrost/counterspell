package counterspell

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-github-username/counterspell/internal/db"
)

func setupLogTestDB(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the logs table
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
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return database
}

func TestSQLiteLogWriter_WriteValidJSON(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)
	defer writer.Close()

	// Test writing a valid JSON log with trace context
	logJSON := `{
		"level": "info",
		"message": "Test log message",
		"timestamp": "2024-01-01T12:00:00Z",
		"trace_id": "1234567890abcdef1234567890abcdef",
		"span_id": "1234567890abcdef",
		"service": "test-service",
		"user_id": "user123"
	}`

	n, err := writer.Write([]byte(logJSON))
	if err != nil {
		t.Fatalf("Failed to write log: %v", err)
	}

	if n != len(logJSON) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(logJSON), n)
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log was written to database
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]
	if log.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", log.Level)
	}
	if log.Message != "Test log message" {
		t.Errorf("Expected message 'Test log message', got '%s'", log.Message)
	}
	if !log.TraceID.Valid || log.TraceID.String != "1234567890abcdef1234567890abcdef" {
		t.Errorf("Expected trace_id '1234567890abcdef1234567890abcdef', got '%s'", log.TraceID.String)
	}
	if !log.SpanID.Valid || log.SpanID.String != "1234567890abcdef" {
		t.Errorf("Expected span_id '1234567890abcdef', got '%s'", log.SpanID.String)
	}

	// Check attributes were stored (excluding processed fields)
	expectedAttrs := `{"service":"test-service","user_id":"user123"}`
	if log.Attributes.String != expectedAttrs {
		t.Errorf("Expected attributes '%s', got '%s'", expectedAttrs, log.Attributes.String)
	}
}

func TestSQLiteLogWriter_WriteWithoutTraceContext(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)
	defer writer.Close()

	// Test writing a log without trace context
	logJSON := `{
		"level": "error",
		"message": "Error occurred",
		"timestamp": "2024-01-01T12:00:00Z",
		"service": "test-service",
		"error_code": "E001"
	}`

	_, err := writer.Write([]byte(logJSON))
	if err != nil {
		t.Fatalf("Failed to write log: %v", err)
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log was written to database
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]
	if log.Level != "error" {
		t.Errorf("Expected level 'error', got '%s'", log.Level)
	}
	if log.Message != "Error occurred" {
		t.Errorf("Expected message 'Error occurred', got '%s'", log.Message)
	}
	if log.TraceID.Valid {
		t.Errorf("Expected trace_id to be NULL, got '%s'", log.TraceID.String)
	}
	if log.SpanID.Valid {
		t.Errorf("Expected span_id to be NULL, got '%s'", log.SpanID.String)
	}

	// Check attributes were stored
	expectedAttrs := `{"error_code":"E001","service":"test-service"}`
	if log.Attributes.String != expectedAttrs {
		t.Errorf("Expected attributes '%s', got '%s'", expectedAttrs, log.Attributes.String)
	}
}

func TestSQLiteLogWriter_WriteInvalidJSON(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)
	defer writer.Close()

	// Test writing invalid JSON (should be handled gracefully)
	invalidJSON := `{invalid json}`

	n, err := writer.Write([]byte(invalidJSON))
	// Should not return error, but should handle gracefully
	if err != nil {
		t.Fatalf("Writer should handle invalid JSON gracefully, got error: %v", err)
	}

	if n != len(invalidJSON) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(invalidJSON), n)
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify no logs were written to database (invalid JSON is discarded)
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected no logs for invalid JSON, got %d", len(logs))
	}
}

func TestSQLiteLogWriter_WriteMissingFields(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)
	defer writer.Close()

	// Test writing JSON with missing required fields
	logJSON := `{
		"timestamp": "2024-01-01T12:00:00Z",
		"custom_field": "custom_value"
	}`

	_, err := writer.Write([]byte(logJSON))
	if err != nil {
		t.Fatalf("Failed to write log: %v", err)
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log was written with default values
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]
	// Should have default values for missing fields
	if log.Level == "" {
		t.Errorf("Expected non-empty level, got empty string")
	}
	if log.Message == "" {
		t.Errorf("Expected non-empty message, got empty string")
	}

	// Custom field should be in attributes
	expectedAttrs := `{"custom_field":"custom_value"}`
	if log.Attributes.String != expectedAttrs {
		t.Errorf("Expected attributes '%s', got '%s'", expectedAttrs, log.Attributes.String)
	}
}

func TestSQLiteLogWriter_BatchProcessing(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)
	defer writer.Close()

	// Write many logs to test batching
	numLogs := 150 // More than batch size (100)
	for i := 0; i < numLogs; i++ {
		logJSON := `{
			"level": "info",
			"message": "Batch test log ` + string(rune('0'+i%10)) + `",
			"timestamp": "2024-01-01T12:00:00Z",
			"batch_id": "` + string(rune('0'+i%10)) + `"
		}`

		_, err := writer.Write([]byte(logJSON))
		if err != nil {
			t.Errorf("Failed to write log %d: %v", i, err)
		}
	}

	// Wait for async processing to complete
	time.Sleep(1 * time.Second)

	// Verify all logs were processed
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  int64(numLogs + 10), // Get more than we wrote
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	if len(logs) != numLogs {
		t.Errorf("Expected %d logs, got %d", numLogs, len(logs))
	}
}

func TestSQLiteLogWriter_ConcurrentWrites(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)
	defer writer.Close()

	// Test concurrent writes
	const numGoroutines = 10
	const logsPerGoroutine = 20

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < logsPerGoroutine; j++ {
				logJSON := `{
					"level": "info",
					"message": "Concurrent test log",
					"timestamp": "2024-01-01T12:00:00Z",
					"routine_id": "` + string(rune('0'+routineID%10)) + `",
					"log_id": "` + string(rune('0'+j%10)) + `"
				}`

				_, err := writer.Write([]byte(logJSON))
				if err != nil {
					t.Errorf("Goroutine %d failed to write log %d: %v", routineID, j, err)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Verify all logs were processed
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  int64(numGoroutines*logsPerGoroutine + 10),
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	expectedTotal := numGoroutines * logsPerGoroutine
	if len(logs) != expectedTotal {
		t.Errorf("Expected %d logs, got %d", expectedTotal, len(logs))
	}
}

func TestSQLiteLogWriter_Close(t *testing.T) {
	database := setupLogTestDB(t)
	defer database.Close()

	writer := NewSQLiteLogWriter(database)

	// Write a log before closing with unique identifier
	logJSON := `{
		"level": "info",
		"message": "Pre-close log unique test",
		"timestamp": "2024-01-01T12:00:00Z",
		"test_id": "close_test_unique"
	}`

	_, err := writer.Write([]byte(logJSON))
	if err != nil {
		t.Fatalf("Failed to write log: %v", err)
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Verify the log was processed before close
	queries := db.New(database)
	logs, err := queries.GetLogs(context.Background(), db.GetLogsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	// Find our specific log (there might be leftover logs from other operations)
	found := false
	for _, log := range logs {
		if log.Message == "Pre-close log unique test" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find the pre-close log, but it wasn't found")
	}

	// Writing after close should return error
	_, err = writer.Write([]byte(logJSON))
	if err == nil {
		t.Errorf("Expected error when writing to closed writer, got nil")
	}
}
