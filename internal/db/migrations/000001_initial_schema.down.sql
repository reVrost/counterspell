-- Rollback initial schema
DROP INDEX IF EXISTS idx_repo_cache_favorite;
DROP INDEX IF EXISTS idx_repo_cache_user;
DROP INDEX IF EXISTS idx_projects_user;
DROP INDEX IF EXISTS idx_github_connections_user;
DROP INDEX IF EXISTS idx_tasks_project;
DROP INDEX IF EXISTS idx_tasks_user_project;
DROP INDEX IF EXISTS idx_tasks_user_status;

DROP TABLE IF EXISTS repo_cache;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS github_connections;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
