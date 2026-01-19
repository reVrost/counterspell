-- name: GetSettings :one
SELECT openrouter_key, zai_key, anthropic_key, openai_key,
       COALESCE(agent_backend, 'native') as agent_backend, updated_at
FROM settings WHERE id = 1;

-- name: UpsertSettings :exec
UPDATE settings SET
openrouter_key = ?,
zai_key = ?,
anthropic_key = ?,
openai_key = ?,
agent_backend = ?,
updated_at = ?
WHERE id = 1;
