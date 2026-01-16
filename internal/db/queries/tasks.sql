-- name: CreateTask :one
INSERT INTO tasks (id, project_id, title, intent, status, position, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE status = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByProject :many
SELECT * FROM tasks
WHERE project_id = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatusAndProject :many
SELECT * FROM tasks
WHERE status = ? AND project_id = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?;

-- name: UpdateTaskPosition :exec
UPDATE tasks SET position = ?, updated_at = ? WHERE id = ?;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks SET status = ?, position = ?, updated_at = ? WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- name: UpdateTaskResult :exec
UPDATE tasks SET status = ?, agent_output = ?, git_diff = ?, message_history = ?, updated_at = ? WHERE id = ?;

-- name: ClearTaskHistory :exec
UPDATE tasks SET message_history = '', agent_output = '', updated_at = ? WHERE id = ?;

-- name: ResetCompletedInProgressTasks :execresult
UPDATE tasks SET status = ?, updated_at = ?
WHERE status = ? AND (agent_output IS NOT NULL AND agent_output != '');

-- name: ResetStuckInProgressTasks :execresult
UPDATE tasks SET status = ?, updated_at = ?
WHERE status = ? AND (agent_output IS NULL OR agent_output = '');
