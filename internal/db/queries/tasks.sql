-- name: CreateTask :exec
INSERT INTO tasks (id, machine_id, title, intent, status, position, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE status = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByMachine :many
SELECT * FROM tasks
WHERE machine_id = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatusAndMachine :many
SELECT * FROM tasks
WHERE status = ? AND machine_id = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?;

-- name: UpdateTaskStep :exec
UPDATE tasks SET current_step = ?, updated_at = ? WHERE id = ?;

-- name: UpdateTaskStatusAndStep :exec
UPDATE tasks SET status = ?, current_step = ?, updated_at = ? WHERE id = ?;

-- name: UpdateTaskPosition :exec
UPDATE tasks SET position = ?, updated_at = ? WHERE id = ?;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks SET status = ?, position = ?, updated_at = ? WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- name: AssignAgent :exec
UPDATE tasks SET assigned_agent_id = ?, updated_at = ? WHERE id = ?;

-- name: ListTasksByAssignedAgent :many
SELECT * FROM tasks
WHERE assigned_agent_id = ?
ORDER BY created_at DESC;
