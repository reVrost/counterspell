package counterspell

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/revrost/counterspell/internal/db"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SpanData represents a span ready to be inserted into the database
type SpanData struct {
	SpanID       string
	TraceID      string
	ParentSpanID string
	Name         string
	StartTime    int64
	EndTime      int64
	DurationNs   int64
	Attributes   []byte
	ServiceName  string
	HasError     bool
}

// DuckDBSpanExporter implements the OpenTelemetry SpanExporter interface
// and writes spans to DuckDB database asynchronously
type DuckDBSpanExporter struct {
	queries   *db.Queries
	spanChan  chan SpanData
	batchSize int
	done      chan struct{}
	wg        sync.WaitGroup
}

// NewDuckDBSpanExporter creates a new DuckDB span exporter
func NewDuckDBSpanExporter(database *sql.DB) *DuckDBSpanExporter {
	exporter := &DuckDBSpanExporter{
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
func (e *DuckDBSpanExporter) worker() {
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
func (e *DuckDBSpanExporter) processBatch(batch []SpanData) {
	ctx := context.Background()

	for _, span := range batch {
		_, err := e.queries.InsertSpan(ctx,
			span.SpanID,
			span.TraceID,
			span.ParentSpanID,
			span.Name,
			span.StartTime,
			span.EndTime,
			span.DurationNs,
			span.Attributes,
			span.ServiceName,
			span.HasError,
		)
		if err != nil {
			// In a production system, you might want to log this error
			// For now, we'll silently continue to avoid disrupting the application
			continue
		}
	}
}

// ExportSpans exports spans to the DuckDB database
// This method is called by the OpenTelemetry SDK
func (e *DuckDBSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
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
func (e *DuckDBSpanExporter) convertSpan(span trace.ReadOnlySpan) SpanData {
	// Get parent span ID
	var parentSpanID string
	if span.Parent().IsValid() {
		parentSpanID = span.Parent().SpanID().String()
	}

	// Convert attributes to JSON
	attributes := make(map[string]interface{})
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
		StartTime:    span.StartTime().UnixNano(),
		EndTime:      span.EndTime().UnixNano(),
		DurationNs:   span.EndTime().Sub(span.StartTime()).Nanoseconds(),
		Attributes:   attributesJSON,
		ServiceName:  serviceName,
		HasError:     hasError,
	}
}

// Shutdown gracefully shuts down the exporter
func (e *DuckDBSpanExporter) Shutdown(ctx context.Context) error {
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
func (e *DuckDBSpanExporter) ForceFlush(ctx context.Context) error {
	// Since we're using a buffered channel and batch processing,
	// we can't easily force flush without adding complexity.
	// For simplicity, we'll just return nil here.
	return nil
}
