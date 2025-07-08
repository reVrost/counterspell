-- name: InsertSpan :exec
INSERT INTO spans (
  span_id, trace_id, parent_span_id, name, start_time, end_time, duration_ns, attributes, service_name, has_error
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: InsertLog :exec
INSERT INTO logs (
  timestamp, level, message, trace_id, span_id, attributes
) VALUES (
  ?, ?, ?, ?, ?, ?
);

-- name: GetTraceDetails :many
SELECT * FROM spans
WHERE trace_id = ?
ORDER BY start_time ASC;

-- name: GetLogs :many
SELECT * FROM logs
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetLogsWithFilters :many
SELECT * FROM logs
WHERE 
  (? IS NULL OR level = ?)
  AND (? IS NULL OR trace_id = ?)
  AND (? IS NULL OR timestamp >= ?)
  AND (? IS NULL OR timestamp <= ?)
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: CountLogs :one
SELECT COUNT(*) FROM logs;

-- name: CountLogsWithFilters :one
SELECT COUNT(*) FROM logs
WHERE 
  (? IS NULL OR level = ?)
  AND (? IS NULL OR trace_id = ?)
  AND (? IS NULL OR timestamp >= ?)
  AND (? IS NULL OR timestamp <= ?);

-- name: GetRootSpans :many
SELECT trace_id, name, start_time, end_time
FROM spans
WHERE parent_span_id IS NULL OR parent_span_id = ''
ORDER BY start_time DESC
LIMIT ? OFFSET ?;

-- name: GetTraceStats :many
SELECT 
  trace_id,
  COUNT(*) as span_count,
  SUM(CASE WHEN has_error THEN 1 ELSE 0 END) as error_count
FROM spans
GROUP BY trace_id;

-- name: CountTraces :one
SELECT COUNT(DISTINCT trace_id) FROM spans
WHERE parent_span_id IS NULL OR parent_span_id = ''; 