package counterspell

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/revrost/counterspell/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use the same schema setup as the main application
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return database
}

func TestDuckDBSpanExporter_ProcessSpanData(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	exporter := NewDuckDBSpanExporter(database)
	defer exporter.Shutdown(context.Background())

	// Test direct span data processing
	spanData := SpanData{
		SpanID:       "1234567890abcdef",
		TraceID:      "1234567890abcdef1234567890abcdef",
		ParentSpanID: "",
		Name:         "test-span",
		StartTime:    time.Now().UnixNano(),
		EndTime:      time.Now().Add(100 * time.Millisecond).UnixNano(),
		DurationNs:   100000000,
		Attributes:   []byte(`{"http.method":"GET","http.url":"/test"}`),
		ServiceName:  "test-service",
		HasError:     false,
	}

	// Send span data to the channel
	select {
	case exporter.spanChan <- spanData:
	default:
		t.Fatal("Failed to send span data to channel")
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify span was written to database
	queries := db.New(database)
	dbSpans, err := queries.GetTraceDetails(context.Background(), "1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to query spans: %v", err)
	}

	if len(dbSpans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(dbSpans))
	}

	span := dbSpans[0]
	if span.Name != "test-span" {
		t.Errorf("Expected span name 'test-span', got '%s'", span.Name)
	}
	if span.HasError {
		t.Errorf("Expected span to not have error, but it does")
	}
	if span.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", span.ServiceName)
	}

	// Check attributes were stored correctly
	var attrs map[string]interface{}
	if err := json.Unmarshal(span.Attributes, &attrs); err != nil {
		t.Errorf("Failed to parse attributes JSON: %v", err)
	}
	if attrs["http.method"] != "GET" {
		t.Errorf("Expected http.method to be 'GET', got '%v'", attrs["http.method"])
	}
}

func TestDuckDBSpanExporter_BatchProcessing(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	exporter := NewDuckDBSpanExporter(database)
	defer exporter.Shutdown(context.Background())

	// Send many spans to test batching
	numSpans := 150 // More than batch size (100)
	for i := 0; i < numSpans; i++ {
		spanData := SpanData{
			SpanID:       "span" + fmt.Sprintf("%04d", i), // Unique span IDs
			TraceID:      "1234567890abcdef1234567890abcdef",
			ParentSpanID: "",
			Name:         "batch-test-span",
			StartTime:    time.Now().UnixNano(),
			EndTime:      time.Now().Add(100 * time.Millisecond).UnixNano(),
			DurationNs:   100000000,
			Attributes:   []byte(`{"test":"batch"}`),
			ServiceName:  "test-service",
			HasError:     false,
		}

		select {
		case exporter.spanChan <- spanData:
		default:
			t.Errorf("Failed to send span data %d to channel", i)
		}
	}

	// Wait for async processing to complete
	time.Sleep(1 * time.Second)

	// Verify all spans were processed
	queries := db.New(database)
	dbSpans, err := queries.GetTraceDetails(context.Background(), "1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to query spans: %v", err)
	}

	if len(dbSpans) != numSpans {
		t.Errorf("Expected %d spans, got %d", numSpans, len(dbSpans))
	}
}

func TestDuckDBSpanExporter_ErrorSpan(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	exporter := NewDuckDBSpanExporter(database)
	defer exporter.Shutdown(context.Background())

	// Test span with error
	spanData := SpanData{
		SpanID:       "error-span",
		TraceID:      "error-trace",
		ParentSpanID: "",
		Name:         "error-span",
		StartTime:    time.Now().UnixNano(),
		EndTime:      time.Now().Add(100 * time.Millisecond).UnixNano(),
		DurationNs:   100000000,
		Attributes:   []byte(`{"error":"test error"}`),
		ServiceName:  "test-service",
		HasError:     true,
	}

	select {
	case exporter.spanChan <- spanData:
	default:
		t.Fatal("Failed to send error span data to channel")
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify error span was written correctly
	queries := db.New(database)
	dbSpans, err := queries.GetTraceDetails(context.Background(), "error-trace")
	if err != nil {
		t.Fatalf("Failed to query spans: %v", err)
	}

	if len(dbSpans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(dbSpans))
	}

	span := dbSpans[0]
	if !span.HasError {
		t.Errorf("Expected span to have error, but it doesn't")
	}
}

func TestDuckDBSpanExporter_Shutdown(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	exporter := NewDuckDBSpanExporter(database)

	// Add some spans to the queue
	spanData := SpanData{
		SpanID:       "shutdown-span",
		TraceID:      "shutdown-trace",
		ParentSpanID: "",
		Name:         "shutdown-test",
		StartTime:    time.Now().UnixNano(),
		EndTime:      time.Now().Add(100 * time.Millisecond).UnixNano(),
		DurationNs:   100000000,
		Attributes:   []byte(`{"test":"shutdown"}`),
		ServiceName:  "test-service",
		HasError:     false,
	}

	select {
	case exporter.spanChan <- spanData:
	default:
		t.Fatal("Failed to send span data to channel")
	}

	// Test shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := exporter.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown exporter: %v", err)
	}

	// Verify the span was processed before shutdown
	queries := db.New(database)
	dbSpans, err := queries.GetTraceDetails(context.Background(), "shutdown-trace")
	if err != nil {
		t.Fatalf("Failed to query spans: %v", err)
	}

	if len(dbSpans) != 1 {
		t.Errorf("Expected 1 span to be processed during shutdown, got %d", len(dbSpans))
	}
}

func TestDuckDBSpanExporter_ShutdownTimeout(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	exporter := NewDuckDBSpanExporter(database)

	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Shutdown should return context.DeadlineExceeded
	err := exporter.Shutdown(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestDuckDBSpanExporter_ConcurrentProcessing(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	exporter := NewDuckDBSpanExporter(database)
	defer exporter.Shutdown(context.Background())

	// Test concurrent span processing
	const numGoroutines = 10
	const spansPerGoroutine = 20

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < spansPerGoroutine; j++ {
				spanData := SpanData{
					SpanID:       fmt.Sprintf("span-%d-%d", routineID, j), // Unique span IDs
					TraceID:      "concurrent-trace",
					ParentSpanID: "",
					Name:         "concurrent-test",
					StartTime:    time.Now().UnixNano(),
					EndTime:      time.Now().Add(100 * time.Millisecond).UnixNano(),
					DurationNs:   100000000,
					Attributes:   []byte(`{"routine":"` + fmt.Sprintf("%d", routineID) + `"}`),
					ServiceName:  "test-service",
					HasError:     false,
				}

				select {
				case exporter.spanChan <- spanData:
				case <-time.After(1 * time.Second):
					t.Errorf("Timeout sending span data from goroutine %d", routineID)
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

	// Verify all spans were processed
	queries := db.New(database)
	dbSpans, err := queries.GetTraceDetails(context.Background(), "concurrent-trace")
	if err != nil {
		t.Fatalf("Failed to query spans: %v", err)
	}

	expectedTotal := numGoroutines * spansPerGoroutine
	if len(dbSpans) != expectedTotal {
		t.Errorf("Expected %d spans, got %d", expectedTotal, len(dbSpans))
	}
}
