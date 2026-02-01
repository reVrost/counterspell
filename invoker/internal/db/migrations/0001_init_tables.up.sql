-- Base schema (single squashed migration)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Shared trigger function for updated_at timestamps (milliseconds since epoch)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = EXTRACT(EPOCH FROM NOW()) * 1000;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Profiles table (extends Supabase auth.users)
CREATE TABLE IF NOT EXISTS profiles (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    tier TEXT NOT NULL DEFAULT 'free' CHECK(tier IN ('free', 'pro', 'enterprise')),
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000),
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000),
    CHECK(length(username) >= 3 AND length(username) <= 20)
);

CREATE INDEX IF NOT EXISTS idx_profiles_email ON profiles(email);
CREATE INDEX IF NOT EXISTS idx_profiles_username ON profiles(username);
CREATE INDEX IF NOT EXISTS idx_profiles_tier ON profiles(tier);

DROP TRIGGER IF EXISTS update_profiles_timestamp ON profiles;
CREATE TRIGGER update_profiles_timestamp
    BEFORE UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Machine registry (track all user VMs and their states)
CREATE TABLE IF NOT EXISTS machine_registry (
    id TEXT PRIMARY KEY,
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    fly_machine_id TEXT NOT NULL UNIQUE,
    fly_app_name TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('creating', 'running', 'stopped', 'error')),
    subdomain TEXT NOT NULL UNIQUE,
    public_url TEXT NOT NULL,
    region TEXT NOT NULL,
    vm_size TEXT NOT NULL DEFAULT 'shared-cpu-1x',
    volume_id TEXT,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000),
    last_seen_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000),
    last_heartbeat_at BIGINT,
    error_message TEXT
);

-- Waitlist table for capturing early signups
CREATE TABLE IF NOT EXISTS waitlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Pending OAuth logins (for PKCE flow)
CREATE TABLE IF NOT EXISTS pending_oauth_logins (
    id TEXT PRIMARY KEY,
    state TEXT NOT NULL UNIQUE,
    code_challenge TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    auth_code TEXT NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000),
    expires_at BIGINT NOT NULL
);

-- Machine auth (store machine-scoped JWTs)
CREATE TABLE IF NOT EXISTS machine_auth (
    id TEXT PRIMARY KEY,
    machine_id TEXT NOT NULL UNIQUE,
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    subdomain TEXT NOT NULL UNIQUE,
    tunnel_provider TEXT NOT NULL DEFAULT 'cloudflare',
    tunnel_token TEXT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000),
    last_seen_at BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
