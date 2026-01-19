-- name: CreateMessage :exec
INSERT INTO messages (id, task_id, run_id, role, content, tool_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetMessage :one
SELECT * FROM messages WHERE id = ?;

-- name: GetMessagesByTask :many
SELECT * FROM messages WHERE task_id = ? ORDER BY created_at ASC;

-- name: GetMessagesByRun :many
SELECT * FROM messages WHERE run_id = ? ORDER BY created_at ASC;

-- name: GetRecentMessages :many
SELECT * FROM messages WHERE task_id = ? ORDER BY created_at DESC LIMIT ?;

-- name: DeleteMessagesByTask :exec
DELETE FROM messages WHERE task_id = ?;
