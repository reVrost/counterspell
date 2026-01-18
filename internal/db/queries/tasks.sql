-- name: CreateTask :one
INSERT INTO tasks (id, user_id, project_id, title, intent, status, position, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 AND user_id = $2;

-- name: ListTasks :many
SELECT * FROM tasks
WHERE user_id = $1
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE user_id = $1 AND status = $2
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByProject :many
SELECT * FROM tasks
WHERE user_id = $1 AND project_id = $2
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatusAndProject :many
SELECT * FROM tasks
WHERE user_id = $1 AND status = $2 AND project_id = $3
ORDER BY status ASC, position ASC, created_at DESC;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: UpdateTaskStep :exec
UPDATE tasks SET current_step = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: UpdateTaskStatusAndStep :exec
UPDATE tasks SET status = $1, current_step = $2, updated_at = $3 WHERE id = $4 AND user_id = $5;

-- name: UpdateTaskPosition :exec
UPDATE tasks SET position = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks SET status = $1, position = $2, updated_at = $3 WHERE id = $4 AND user_id = $5;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1 AND user_id = $2;

-- name: AssignAgent :exec
UPDATE tasks SET assigned_agent_id = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: AssignUser :exec
UPDATE tasks SET assigned_user_id = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: ListTasksByAssignedAgent :many
SELECT * FROM tasks
WHERE assigned_agent_id = $1
ORDER BY created_at DESC;

-- name: ListTasksByAssignedUser :many
SELECT * FROM tasks
WHERE assigned_user_id = $1
ORDER BY created_at DESC;
