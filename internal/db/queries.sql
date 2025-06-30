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
  (sqlc.narg('level') IS NULL OR level = sqlc.narg('level'))
  AND (sqlc.narg('trace_id') IS NULL OR trace_id = sqlc.narg('trace_id'))
  AND (sqlc.narg('start_time') IS NULL OR timestamp >= sqlc.narg('start_time'))
  AND (sqlc.narg('end_time') IS NULL OR timestamp <= sqlc.narg('end_time'))
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: CountLogs :one
SELECT COUNT(*) FROM logs;

-- name: CountLogsWithFilters :one
SELECT COUNT(*) FROM logs
WHERE 
  (sqlc.narg('level') IS NULL OR level = sqlc.narg('level'))
  AND (sqlc.narg('trace_id') IS NULL OR trace_id = sqlc.narg('trace_id'))
  AND (sqlc.narg('start_time') IS NULL OR timestamp >= sqlc.narg('start_time'))
  AND (sqlc.narg('end_time') IS NULL OR timestamp <= sqlc.narg('end_time'));

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