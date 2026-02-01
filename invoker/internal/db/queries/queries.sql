-- name: CreateUser :one
INSERT INTO profiles (id, email, first_name, last_name, username, tier, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM profiles
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM profiles
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM profiles
WHERE username = $1;

-- name: UsernameExists :one
SELECT EXISTS(SELECT 1 FROM profiles WHERE username = $1);

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM profiles WHERE email = $1);

-- name: Listprofiles :many
SELECT * FROM profiles
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE profiles
SET email = $2, first_name = $3, last_name = $4, username = $5, tier = $6, updated_at = $7
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM profiles
WHERE id = $1;

-- name: CreateSubscription :one
INSERT INTO subscriptions (id, profile_id, stripe_sub_id, tier, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1;

-- name: GetSubscriptionByUserID :one
SELECT * FROM subscriptions
WHERE profile_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListSubscriptionsByUser :many
SELECT * FROM subscriptions
WHERE profile_id = $1
ORDER BY created_at DESC;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET tier = $2, status = $3, current_period_start = $4, current_period_end = $5, cancel_at_period_end = $6, updated_at = $7
WHERE id = $1
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE id = $1;

-- name: CreateMachineRegistry :one
INSERT INTO machine_registry (id, profile_id, fly_machine_id, fly_app_name, status, subdomain, public_url, region, vm_size, volume_id, created_at, last_seen_at, last_heartbeat_at, error_message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetMachineRegistryByID :one
SELECT * FROM machine_registry
WHERE id = $1;

-- name: GetMachineRegistryByUserID :many
SELECT * FROM machine_registry
WHERE profile_id = $1
ORDER BY created_at DESC;

-- name: GetMachineRegistryBySubdomain :one
SELECT * FROM machine_registry
WHERE subdomain = $1;

-- name: GetMachineRegistryByFlyMachineID :one
SELECT * FROM machine_registry
WHERE fly_machine_id = $1;

-- name: ListMachineRegistry :many
SELECT * FROM machine_registry
ORDER BY created_at DESC;

-- name: UpdateMachineRegistry :one
UPDATE machine_registry
SET status = $2, last_seen_at = $3, last_heartbeat_at = $4, error_message = $5
WHERE id = $1
RETURNING *;

-- name: UpdateMachineRegistryStatus :one
UPDATE machine_registry
SET status = $2, last_seen_at = $3, error_message = $4
WHERE id = $1
RETURNING *;

-- name: DeleteMachineRegistry :exec
DELETE FROM machine_registry
WHERE id = $1;

-- name: CreateRoutingTable :one
INSERT INTO routing_table (subdomain, fly_machine_id, fly_url, is_active, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRoutingTableBySubdomain :one
SELECT * FROM routing_table
WHERE subdomain = $1;

-- name: ListRoutingTable :many
SELECT * FROM routing_table
ORDER BY updated_at DESC;

-- name: UpdateRoutingTable :one
UPDATE routing_table
SET fly_machine_id = $2, fly_url = $3, is_active = $4, updated_at = $5
WHERE subdomain = $1
RETURNING *;

-- name: DeleteRoutingTable :exec
DELETE FROM routing_table
WHERE subdomain = $1;

-- name: CreateUsageTracking :one
INSERT INTO usage_tracking (id, profile_id, machine_id, metric_type, quantity, recorded_at, period_start, period_end)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUsageTrackingByUserIDAndPeriod :many
SELECT * FROM usage_tracking
WHERE profile_id = $1 AND period_start >= $2 AND period_end <= $3
ORDER BY recorded_at DESC;

-- name: ListUsageTracking :many
SELECT * FROM usage_tracking
ORDER BY recorded_at DESC;

-- name: CreateAuditLog :one
INSERT INTO audit_log (id, profile_id, action, resource_type, resource_id, ip_address, user_agent, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_log
WHERE profile_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetQuotaLimitByTier :one
SELECT * FROM quota_limits
WHERE tier = $1;

-- name: ListQuotaLimits :many
SELECT * FROM quota_limits
ORDER BY tier;

-- OAuth pending logins queries
-- name: CreatePendingOAuthLogin :one
INSERT INTO pending_oauth_logins (id, state, code_challenge, redirect_uri, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPendingOAuthLoginByState :one
SELECT * FROM pending_oauth_logins
WHERE state = $1 AND expires_at > EXTRACT(EPOCH FROM NOW()) * 1000;

-- name: UpdatePendingOAuthLoginAuthCode :one
UPDATE pending_oauth_logins
SET auth_code = $2
WHERE state = $1 AND expires_at > EXTRACT(EPOCH FROM NOW()) * 1000
RETURNING *;

-- name: DeletePendingOAuthLogin :exec
DELETE FROM pending_oauth_logins
WHERE state = $1;

-- name: CleanupExpiredOAuthLogins :exec
DELETE FROM pending_oauth_logins
WHERE expires_at <= EXTRACT(EPOCH FROM NOW()) * 1000;

-- Machine auth queries
-- name: CreateMachineAuth :one
INSERT INTO machine_auth (id, machine_id, profile_id, subdomain, tunnel_provider, tunnel_token, created_at, last_seen_at, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetMachineAuthByMachineID :one
SELECT * FROM machine_auth
WHERE machine_id = $1 AND is_active = TRUE;

-- name: GetMachineAuthByUserID :many
SELECT * FROM machine_auth
WHERE profile_id = $1 AND is_active = TRUE
ORDER BY created_at DESC;

-- name: GetMachineAuthBySubdomain :one
SELECT * FROM machine_auth
WHERE subdomain = $1 AND is_active = TRUE;

-- name: RevokeMachineAuth :one
UPDATE machine_auth
SET is_active = FALSE
WHERE machine_id = $1
RETURNING *;

-- name: UpdateMachineAuthLastSeen :one
UPDATE machine_auth
SET last_seen_at = $2
WHERE machine_id = $1
RETURNING *;

-- name: UpdateMachineAuthTunnel :one
UPDATE machine_auth
SET tunnel_provider = $2, tunnel_token = $3, subdomain = $4
WHERE machine_id = $1 AND is_active = TRUE
RETURNING *;
