-- name: GetSettings :one
SELECT openrouter_key, zai_key, anthropic_key, openai_key,
       COALESCE(agent_backend, 'native') as agent_backend,
       COALESCE(provider, 'anthropic') as provider,
       COALESCE(model, 'claude-opus-4-5') as model,
       updated_at
FROM settings WHERE id = 1;

-- name: UpsertSettings :exec
INSERT INTO settings (
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
)
ON CONFLICT(id) DO UPDATE SET
    openrouter_key = excluded.openrouter_key,
    zai_key = excluded.zai_key,
    anthropic_key = excluded.anthropic_key,
    openai_key = excluded.openai_key,
    agent_backend = excluded.agent_backend,
    provider = excluded.provider,
    model = excluded.model,
    updated_at = excluded.updated_at;

-- name: GetMachineJWT :one
SELECT machine_jwt
FROM settings
WHERE id = 1;

-- name: UpdateMachineJWT :exec
UPDATE settings
SET machine_jwt = ?,
    updated_at = ?
WHERE id = 1;

-- name: GetMachineID :one
SELECT machine_id
FROM settings
WHERE id = 1;

-- name: UpdateMachineID :exec
UPDATE settings
SET machine_id = ?,
    updated_at = ?
WHERE id = 1;
