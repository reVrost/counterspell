-- name: GetUserSettings :one
SELECT user_id, openrouter_key, zai_key, anthropic_key, openai_key, 
       COALESCE(agent_backend, 'native') as agent_backend, updated_at 
FROM user_settings WHERE user_id = $1;

-- name: UpsertUserSettings :exec
INSERT INTO user_settings (user_id, openrouter_key, zai_key, anthropic_key, openai_key, agent_backend, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT(user_id) DO UPDATE SET
openrouter_key = EXCLUDED.openrouter_key,
zai_key = EXCLUDED.zai_key,
anthropic_key = EXCLUDED.anthropic_key,
openai_key = EXCLUDED.openai_key,
agent_backend = EXCLUDED.agent_backend,
updated_at = EXCLUDED.updated_at;
