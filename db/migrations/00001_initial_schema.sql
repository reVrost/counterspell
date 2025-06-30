-- +goose Up
-- +goose StatementBegin
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    trace_id TEXT,
    span_id TEXT,
    attributes TEXT -- Stored as a JSON string
);

CREATE TABLE spans (
    span_id TEXT PRIMARY KEY,
    trace_id TEXT NOT NULL,
    parent_span_id TEXT,
    name TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    duration_ns INTEGER NOT NULL,
    attributes TEXT, -- Stored as a JSON string
    service_name TEXT NOT NULL,
    has_error BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_logs_timestamp ON logs(timestamp);
CREATE INDEX idx_logs_trace_id ON logs(trace_id);
CREATE INDEX idx_spans_trace_id ON spans(trace_id);
CREATE INDEX idx_spans_start_time ON spans(start_time);
CREATE INDEX idx_spans_has_error ON spans(has_error);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS spans;
-- +goose StatementEnd 