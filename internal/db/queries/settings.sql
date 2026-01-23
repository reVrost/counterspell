-- name: GetSettings :one
SELECT openrouter_key, zai_key, anthropic_key, openai_key,
       COALESCE(agent_backend, 'native') as agent_backend,
       COALESCE(provider, 'anthropic') as provider,
       COALESCE(model, 'claude-opus-4-5') as model,
       updated_at
FROM settings WHERE id = 1;

-- name: UpsertSettings :exec
INSERT OR REPLACE INTO settings (
    id,
    openrouter_key,
    zai_key,
    anthropic_key,
    openai_key,
    agent_backend,
    provider,
    model,
    updated_at
) VALUES (
    1,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);
