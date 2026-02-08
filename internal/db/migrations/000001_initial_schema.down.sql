-- Counterspell Initial Schema
-- Down migration

DROP TRIGGER IF EXISTS update_tasks_updated_at;
DROP TRIGGER IF EXISTS update_agent_runs_updated_at;
DROP TRIGGER IF EXISTS update_artifacts_updated_at;
DROP TRIGGER IF EXISTS update_messages_updated_at;
DROP TRIGGER IF EXISTS update_settings_updated_at;
DROP TRIGGER IF EXISTS update_workspaces_updated_at;
DROP TRIGGER IF EXISTS update_sessions_updated_at;
DROP TRIGGER IF EXISTS update_session_message_count_on_insert;
DROP TRIGGER IF EXISTS update_session_message_count_on_delete;
DROP TRIGGER IF EXISTS update_run_message_count_on_insert;
DROP TRIGGER IF EXISTS update_run_message_count_on_delete;

DROP TABLE IF EXISTS session_messages;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS github_connections;
DROP TABLE IF EXISTS machine_identity;
DROP TABLE IF EXISTS connector_auth;
DROP TABLE IF EXISTS connector_oauth_attempts;
DROP TABLE IF EXISTS oauth_login_attempts;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS artifacts;
DROP TABLE IF EXISTS agent_runs;
DROP TABLE IF EXISTS tasks;
