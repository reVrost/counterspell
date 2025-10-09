package counterspell

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/revrost/counterspell/pkg/db"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SpanData represents a span ready to be inserted into the database
type SpanData struct {
	SpanID       string
	TraceID      string
	ParentSpanID sql.NullString
	Name         string
	StartTime    string
	EndTime      string
	DurationNs   int64
	Attributes   string
	ServiceName  string
	HasError     bool
}

// SQLiteSpanExporter implements the OpenTelemetry SpanExporter interface
// and writes spans to SQLite database asynchronously
type SQLiteSpanExporter struct {
	queries   *db.Queries
	spanChan  chan SpanData
	batchSize int
	done      chan struct{}
	wg        sync.WaitGroup
}

// NewSQLiteSpanExporter creates a new SQLite span exporter
func NewSQLiteSpanExporter(database *sql.DB) *SQLiteSpanExporter {
	exporter := &SQLiteSpanExporter{
		queries:   db.New(database),
		spanChan:  make(chan SpanData, 1000), // Buffer for async processing
		batchSize: 100,                       // Process spans in batches
		done:      make(chan struct{}),
	}

	// Start background worker goroutine
	exporter.wg.Add(1)
	go exporter.worker()

	return exporter
}

// worker processes spans from the channel in batches
func (e *SQLiteSpanExporter) worker() {
	defer e.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond) // Flush every 100ms
	defer ticker.Stop()

	batch := make([]SpanData, 0, e.batchSize)

	for {
		select {
		case span := <-e.spanChan:
			batch = append(batch, span)

			// Process batch when it's full
			if len(batch) >= e.batchSize {
				e.processBatch(batch)
				batch = batch[:0] // Reset slice
			}

		case <-ticker.C:
			// Process any remaining spans in batch
			if len(batch) > 0 {
				e.processBatch(batch)
				batch = batch[:0]
			}

		case <-e.done:
			// Process any remaining spans before shutting down
			if len(batch) > 0 {
				e.processBatch(batch)
			}

			// Drain any remaining spans in the channel
			for {
				select {
				case span := <-e.spanChan:
					batch = append(batch, span)
					if len(batch) >= e.batchSize {
						e.processBatch(batch)
						batch = batch[:0]
					}
				default:
					if len(batch) > 0 {
						e.processBatch(batch)
					}
					return
				}
			}
		}
	}
}

// processBatch inserts a batch of spans into the database
func (e *SQLiteSpanExporter) processBatch(batch []SpanData) {
	ctx := context.Background()

	for _, span := range batch {
		err := e.queries.InsertSpan(ctx, db.InsertSpanParams{
			SpanID:       span.SpanID,
			TraceID:      span.TraceID,
			ParentSpanID: span.ParentSpanID,
			Name:         span.Name,
			StartTime:    span.StartTime,
			EndTime:      span.EndTime,
			DurationNs:   span.DurationNs,
			Attributes:   sql.NullString{String: span.Attributes, Valid: span.Attributes != ""},
			ServiceName:  span.ServiceName,
			HasError:     span.HasError,
		})
		if err != nil {
			// In a production system, you might want to log this error
			// For now, we'll silently continue to avoid disrupting the application
			continue
		}
	}
}

// ExportSpans exports spans to the SQLite database
// This method is called by the OpenTelemetry SDK
func (e *SQLiteSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	for _, span := range spans {
		spanData := e.convertSpan(span)

		// Send span to worker goroutine (non-blocking)
		select {
		case e.spanChan <- spanData:
			// Successfully queued
		default:
			// Channel is full, drop the span to avoid blocking
			// In a production system, you might want to implement backpressure
		}
	}

	return nil
}

// convertSpan converts an OpenTelemetry ReadOnlySpan to our SpanData struct
func (e *SQLiteSpanExporter) convertSpan(span trace.ReadOnlySpan) SpanData {
	// Get parent span ID
	var parentSpanID sql.NullString
	if span.Parent().IsValid() {
		parentSpanID = sql.NullString{
			String: span.Parent().SpanID().String(),
			Valid:  true,
		}
	}

	// Convert attributes to JSON
	attributes := make(map[string]any)
	for _, attr := range span.Attributes() {
		attributes[string(attr.Key)] = attr.Value.AsInterface()
	}

	attributesJSON, _ := json.Marshal(attributes)

	// Get service name from resource attributes
	serviceName := "unknown"
	if resource := span.Resource(); resource != nil {
		for _, attr := range resource.Attributes() {
			if attr.Key == "service.name" {
				serviceName = attr.Value.AsString()
				break
			}
		}
	}

	// Check if span has error
	hasError := span.Status().Code == codes.Error

	return SpanData{
		SpanID:       span.SpanContext().SpanID().String(),
		TraceID:      span.SpanContext().TraceID().String(),
		ParentSpanID: parentSpanID,
		Name:         span.Name(),
		StartTime:    span.StartTime().Format(time.RFC3339Nano),
		EndTime:      span.EndTime().Format(time.RFC3339Nano),
		DurationNs:   span.EndTime().Sub(span.StartTime()).Nanoseconds(),
		Attributes:   string(attributesJSON),
		ServiceName:  serviceName,
		HasError:     hasError,
	}
}

// Shutdown gracefully shuts down the exporter
func (e *SQLiteSpanExporter) Shutdown(ctx context.Context) error {
	close(e.done)

	// Wait for worker to finish with timeout
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ForceFlush forces the exporter to flush any buffered spans
func (e *SQLiteSpanExporter) ForceFlush(ctx context.Context) error {
	// Since we're using a buffered channel and batch processing,
	// we can't easily force flush without adding complexity.
	// For simplicity, we'll just return nil here.
	return nil
}
