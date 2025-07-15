package counterspell

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/revrost/counterspell/internal/db"
)

// LogData represents a log entry for database insertion
type LogData struct {
	Timestamp  int64
	Level      string
	Message    string
	TraceID    string
	SpanID     string
	Attributes []byte
}

// DuckDBLogWriter writes logs to DuckDB asynchronously
type DuckDBLogWriter struct {
	queries   *db.Queries
	logChan   chan LogData
	batchSize int
	done      chan struct{}
	wg        sync.WaitGroup
	closed    bool
	closeMu   sync.RWMutex
}

// NewDuckDBLogWriter creates a new log writer
func NewDuckDBLogWriter(database *sql.DB) *DuckDBLogWriter {
	writer := &DuckDBLogWriter{
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
func (w *DuckDBLogWriter) Write(p []byte) (n int, err error) {
	w.closeMu.RLock()
	if w.closed {
		w.closeMu.RUnlock()
		return 0, sql.ErrConnDone
	}
	w.closeMu.RUnlock()

	var logEntry map[string]any
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
func (w *DuckDBLogWriter) worker() {
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
func (w *DuckDBLogWriter) processBatch(batch []LogData) {
	ctx := context.Background()

	for _, log := range batch {
		_, _ = w.queries.InsertLog(ctx,
			log.Timestamp,
			log.Level,
			log.Message,
			log.TraceID,
			log.SpanID,
			log.Attributes,
		)
	}
}

// convertLogEntry converts JSON log to LogData
func (w *DuckDBLogWriter) convertLogEntry(entry map[string]any) LogData {
	logData := LogData{}

	// Check both "time" and "timestamp" fields
	timestampStr := ""
	if ts, ok := entry["time"].(string); ok {
		timestampStr = ts
	} else if ts, ok := entry["timestamp"].(string); ok {
		timestampStr = ts
	}

	// Parse timestamp string to int64 (nanoseconds)
	if timestampStr != "" {
		if t, err := time.Parse(time.RFC3339Nano, timestampStr); err == nil {
			logData.Timestamp = t.UnixNano()
		}
	} else {
		logData.Timestamp = time.Now().UnixNano()
	}

	if level, ok := entry["level"].(string); ok {
		logData.Level = level
	} else {
		logData.Level = "info"
	}

	if message, ok := entry["message"].(string); ok {
		logData.Message = message
	} else {
		logData.Message = "unknown"
	}

	if traceID, ok := entry["trace_id"].(string); ok {
		logData.TraceID = traceID
	}

	if spanID, ok := entry["span_id"].(string); ok {
		logData.SpanID = spanID
	}

	attributes := make(map[string]any)
	for key, value := range entry {
		if key != "time" && key != "timestamp" && key != "level" && key != "message" && key != "trace_id" && key != "span_id" {
			attributes[key] = value
		}
	}

	if len(attributes) > 0 {
		if attributesJSON, err := json.Marshal(attributes); err == nil {
			logData.Attributes = attributesJSON
		}
	}

	return logData
}

// Close shuts down the writer
func (w *DuckDBLogWriter) Close() error {
	w.closeMu.Lock()
	if w.closed {
		w.closeMu.Unlock()
		return nil
	}
	w.closed = true
	w.closeMu.Unlock()

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
