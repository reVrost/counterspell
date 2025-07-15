package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb/v2"
)

// DBTX is an interface that wraps the basic database operations.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Queries provides a way to execute SQL queries.
type Queries struct {
	db DBTX
}

// New creates a new Queries object.
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// Open opens a new DuckDB database connection.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		return nil, err
	}

	// Manually apply initial schema
	schemaSQL := `
	CREATE SEQUENCE IF NOT EXISTS logs_id_seq;

	CREATE TABLE IF NOT EXISTS spans (
		span_id VARCHAR NOT NULL,
		trace_id VARCHAR NOT NULL,
		parent_span_id VARCHAR,
		name VARCHAR NOT NULL,
		start_time BIGINT NOT NULL,
		end_time BIGINT NOT NULL,
		duration_ns BIGINT NOT NULL,
		attributes BLOB,
		service_name VARCHAR NOT NULL,
		has_error BOOLEAN NOT NULL,
		PRIMARY KEY (span_id, trace_id)
	);

	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY DEFAULT nextval('logs_id_seq'),
		timestamp BIGINT NOT NULL,
		level VARCHAR NOT NULL,
		message VARCHAR NOT NULL,
		trace_id VARCHAR,
		span_id VARCHAR,
		attributes BLOB
	);
	`
	_, err = db.ExecContext(context.Background(), schemaSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to apply initial schema: %w", err)
	}

	return db, nil
}

const insertSpan = `
INSERT INTO spans (
  span_id, trace_id, parent_span_id, name, start_time, end_time, duration_ns, attributes, service_name, has_error
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);
`

func (q *Queries) InsertSpan(ctx context.Context,
	spanID, traceID, parentSpanID, name string,
	startTime, endTime int64,
	durationNs int64,
	attributes []byte,
	serviceName string,
	hasError bool) (sql.Result, error) {
	return q.db.ExecContext(ctx, insertSpan,
		spanID,
		traceID,
		parentSpanID,
		name,
		startTime,
		endTime,
		durationNs,
		attributes,
		serviceName,
		hasError,
	)
}

const insertLog = `
INSERT INTO logs (
  timestamp, level, message, trace_id, span_id, attributes
) VALUES (
  ?, ?, ?, ?, ?, ?
);
`

func (q *Queries) InsertLog(ctx context.Context,
	timestamp int64,
	level, message, traceID, spanID string,
	attributes []byte) (sql.Result, error) {
	return q.db.ExecContext(ctx, insertLog,
		timestamp,
		level,
		message,
		traceID,
		spanID,
		attributes,
	)
}

const getTraceDetails = `
SELECT span_id, trace_id, parent_span_id, name, start_time, end_time, duration_ns, attributes, service_name, has_error FROM spans
WHERE trace_id = ?
ORDER BY start_time ASC;
`

func (q *Queries) GetTraceDetails(ctx context.Context, traceID string) ([]Span, error) {
	rows, err := q.db.QueryContext(ctx, getTraceDetails, traceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Span
	for rows.Next() {
		var i Span
		var parentSpanID sql.NullString
		var attributes sql.NullString // Use sql.NullString for scanning, then convert to []byte
		err := rows.Scan(
			&i.SpanID,
			&i.TraceID,
			&parentSpanID,
			&i.Name,
			&i.StartTime,
			&i.EndTime,
			&i.DurationNs,
			&attributes,
			&i.ServiceName,
			&i.HasError,
		)
		if err != nil {
			return nil, err
		}
		if parentSpanID.Valid {
			i.ParentSpanID = parentSpanID.String
		}
		if attributes.Valid {
			i.Attributes = []byte(attributes.String)
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLogs = `
SELECT id, timestamp, level, message, trace_id, span_id, attributes FROM logs
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;
`

func (q *Queries) GetLogs(ctx context.Context, limit, offset int32) ([]Log, error) {
	rows, err := q.db.QueryContext(ctx, getLogs, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Log
	for rows.Next() {
		var i Log
		var traceID sql.NullString
		var spanID sql.NullString
		var attributes sql.NullString
		err := rows.Scan(
			&i.ID,
			&i.Timestamp,
			&i.Level,
			&i.Message,
			&traceID,
			&spanID,
			&attributes,
		)
		if err != nil {
			return nil, err
		}
		if traceID.Valid {
			i.TraceID = traceID.String
		}
		if spanID.Valid {
			i.SpanID = spanID.String
		}
		if attributes.Valid {
			i.Attributes = []byte(attributes.String)
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLogsWithFilters = `
SELECT id, timestamp, level, message, trace_id, span_id, attributes FROM logs
WHERE 
  (? IS NULL OR level = ?)
  AND (? IS NULL OR trace_id = ?)
  AND (? IS NULL OR timestamp >= ?)
  AND (? IS NULL OR timestamp <= ?)
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;
`

func (q *Queries) GetLogsWithFilters(ctx context.Context,
	level sql.NullString,
	traceID sql.NullString,
	startTime sql.NullInt64,
	endTime sql.NullInt64,
	limit, offset int32) ([]Log, error) {
	rows, err := q.db.QueryContext(ctx, getLogsWithFilters,
		level, level,
		traceID, traceID,
		startTime, startTime,
		endTime, endTime,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Log
	for rows.Next() {
		var i Log
		var logTraceID sql.NullString
		var logSpanID sql.NullString
		var logAttributes sql.NullString
		err := rows.Scan(
			&i.ID,
			&i.Timestamp,
			&i.Level,
			&i.Message,
			&logTraceID,
			&logSpanID,
			&logAttributes,
		)
		if err != nil {
			return nil, err
		}
		if logTraceID.Valid {
			i.TraceID = logTraceID.String
		}
		if logSpanID.Valid {
			i.SpanID = logSpanID.String
		}
		if logAttributes.Valid {
			i.Attributes = []byte(logAttributes.String)
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const countLogs = `
SELECT COUNT(*) FROM logs;
`

func (q *Queries) CountLogs(ctx context.Context) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, countLogs).Scan(&count)
	return count, err
}

const countLogsWithFilters = `
SELECT COUNT(*) FROM logs
WHERE
  (? IS NULL OR level = ?)
  AND (? IS NULL OR trace_id = ?)
  AND (? IS NULL OR timestamp >= ?)
  AND (? IS NULL OR timestamp <= ?);
`

func (q *Queries) CountLogsWithFilters(ctx context.Context,
	level sql.NullString,
	traceID sql.NullString,
	startTime sql.NullInt64,
	endTime sql.NullInt64) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, countLogsWithFilters,
		level, level,
		traceID, traceID,
		startTime, startTime,
		endTime, endTime,
	).Scan(&count)
	return count, err
}

const getRootSpans = `
SELECT trace_id, name, start_time, end_time
FROM spans
WHERE parent_span_id IS NULL OR parent_span_id = ''
ORDER BY start_time DESC
LIMIT ? OFFSET ?;
`

type GetRootSpansRow struct {
	TraceID   string
	Name      string
	StartTime int64
	EndTime   int64
}

func (q *Queries) GetRootSpans(ctx context.Context, limit, offset int32) ([]GetRootSpansRow, error) {
	rows, err := q.db.QueryContext(ctx, getRootSpans, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRootSpansRow
	for rows.Next() {
		var i GetRootSpansRow
		err := rows.Scan(
			&i.TraceID,
			&i.Name,
			&i.StartTime,
			&i.EndTime,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTraceStats = `
SELECT 
  trace_id,
  COUNT(*) as span_count,
  SUM(CASE WHEN has_error THEN 1 ELSE 0 END) as error_count
FROM spans
GROUP BY trace_id;
`

type GetTraceStatsRow struct {
	TraceID    string
	SpanCount  int64
	ErrorCount int64
}

func (q *Queries) GetTraceStats(ctx context.Context) ([]GetTraceStatsRow, error) {
	rows, err := q.db.QueryContext(ctx, getTraceStats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTraceStatsRow
	for rows.Next() {
		var i GetTraceStatsRow
		err := rows.Scan(
			&i.TraceID,
			&i.SpanCount,
			&i.ErrorCount,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const countTraces = `
SELECT COUNT(DISTINCT trace_id) FROM spans
WHERE parent_span_id IS NULL OR parent_span_id = '';
`

func (q *Queries) CountTraces(ctx context.Context) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, countTraces).Scan(&count)
	return count, err
}
