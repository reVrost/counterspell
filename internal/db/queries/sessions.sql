-- Sessions

-- name: CreateSession :exec
INSERT INTO sessions (id, agent_backend, external_id, backend_session_id, title, message_count, last_message_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions WHERE id = ?;

-- name: GetSessionByBackendExternal :one
SELECT * FROM sessions WHERE agent_backend = ? AND external_id = ?;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY COALESCE(last_message_at, created_at) DESC, created_at DESC;

-- name: UpdateSession :exec
UPDATE sessions
SET backend_session_id = ?, title = ?, last_message_at = ?, updated_at = ?
WHERE id = ?;

-- name: UpdateSessionBackendSessionID :exec
UPDATE sessions
SET backend_session_id = ?, updated_at = ?
WHERE id = ?;

-- name: UpdateSessionTitle :exec
UPDATE sessions
SET title = ?, updated_at = ?
WHERE id = ?;

-- name: GetSessionNextSequence :one
SELECT CAST(COALESCE(MAX(sequence) + 1, 0) AS INTEGER) as next_sequence
FROM session_messages
WHERE session_id = ?;

-- name: CreateSessionMessage :exec
INSERT INTO session_messages (id, session_id, sequence, role, kind, content, tool_name, tool_call_id, raw_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListSessionMessages :many
SELECT * FROM session_messages
WHERE session_id = ?
ORDER BY sequence ASC;
