-- name: CreateTask :exec
INSERT INTO tasks (id, repository_id, title, intent, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE status = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = ? WHERE id = ?;

-- name: UpdateTaskPosition :exec
UPDATE tasks SET position = ? WHERE id = ?;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks SET status = ?, position = ? WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;
