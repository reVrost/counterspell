package microscope

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/your-github-username/microscope/internal/db"
)

// LogData represents a log entry for database insertion
type LogData struct {
	Timestamp  string
	Level      string
	Message    string
	TraceID    sql.NullString
	SpanID     sql.NullString
	Attributes string
}

// SQLiteLogWriter writes logs to SQLite asynchronously
type SQLiteLogWriter struct {
	queries   *db.Queries
	logChan   chan LogData
	batchSize int
	done      chan struct{}
	wg        sync.WaitGroup
}

// NewSQLiteLogWriter creates a new log writer
func NewSQLiteLogWriter(database *sql.DB) *SQLiteLogWriter {
	writer := &SQLiteLogWriter{
		queries:   db.New(database),
		logChan:   make(chan LogData, 1000),
		batchSize: 100,
		done:      make(chan struct{}),
	}

	writer.wg.Add(1)
	go writer.worker()

	return writer
}

// Write implements io.Writer
func (w *SQLiteLogWriter) Write(p []byte) (n int, err error) {
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return len(p), nil
	}

	logData := w.convertLogEntry(logEntry)

	select {
	case w.logChan <- logData:
	default:
	}

	return len(p), nil
}

// worker processes logs in batches
func (w *SQLiteLogWriter) worker() {
	defer w.wg.Done()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	batch := make([]LogData, 0, w.batchSize)
	
	for {
		select {
		case log := <-w.logChan:
			batch = append(batch, log)
			if len(batch) >= w.batchSize {
				w.processBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				w.processBatch(batch)
				batch = batch[:0]
			}
		case <-w.done:
			if len(batch) > 0 {
				w.processBatch(batch)
			}
			for {
				select {
				case log := <-w.logChan:
					batch = append(batch, log)
					if len(batch) >= w.batchSize {
						w.processBatch(batch)
						batch = batch[:0]
					}
				default:
					if len(batch) > 0 {
						w.processBatch(batch)
					}
					return
				}
			}
		}
	}
}

// processBatch inserts logs into database
func (w *SQLiteLogWriter) processBatch(batch []LogData) {
	ctx := context.Background()
	
	for _, log := range batch {
		_ = w.queries.InsertLog(ctx, db.InsertLogParams{
			Timestamp:  log.Timestamp,
			Level:      log.Level,
			Message:    log.Message,
			TraceID:    log.TraceID,
			SpanID:     log.SpanID,
			Attributes: sql.NullString{String: log.Attributes, Valid: log.Attributes != ""},
		})
	}
}

// convertLogEntry converts JSON log to LogData
func (w *SQLiteLogWriter) convertLogEntry(entry map[string]interface{}) LogData {
	logData := LogData{}

	if timestamp, ok := entry["time"].(string); ok {
		logData.Timestamp = timestamp
	} else {
		logData.Timestamp = time.Now().Format(time.RFC3339Nano)
	}

	if level, ok := entry["level"].(string); ok {
		logData.Level = level
	} else {
		logData.Level = "info"
	}

	if message, ok := entry["message"].(string); ok {
		logData.Message = message
	}

	if traceID, ok := entry["trace_id"].(string); ok && traceID != "" {
		logData.TraceID = sql.NullString{String: traceID, Valid: true}
	}

	if spanID, ok := entry["span_id"].(string); ok && spanID != "" {
		logData.SpanID = sql.NullString{String: spanID, Valid: true}
	}

	attributes := make(map[string]interface{})
	for key, value := range entry {
		if key != "time" && key != "level" && key != "message" && key != "trace_id" && key != "span_id" {
			attributes[key] = value
		}
	}

	if len(attributes) > 0 {
		if attributesJSON, err := json.Marshal(attributes); err == nil {
			logData.Attributes = string(attributesJSON)
		}
	}

	return logData
}

// Close shuts down the writer
func (w *SQLiteLogWriter) Close() error {
	close(w.done)
	
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return nil
	}
} 