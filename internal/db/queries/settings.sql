-- name: GetUserSettings :one
SELECT user_id, openrouter_key, zai_key, anthropic_key, openai_key, 
       COALESCE(agent_backend, 'native') as agent_backend, updated_at 
FROM user_settings WHERE user_id = 'default';

-- name: UpsertUserSettings :exec
INSERT INTO user_settings (user_id, openrouter_key, zai_key, anthropic_key, openai_key, agent_backend, updated_at)
VALUES ('default', ?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
openrouter_key = excluded.openrouter_key,
zai_key = excluded.zai_key,
anthropic_key = excluded.anthropic_key,
openai_key = excluded.openai_key,
agent_backend = excluded.agent_backend,
updated_at = excluded.updated_at;
