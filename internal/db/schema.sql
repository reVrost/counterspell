-- Enforce WAL mode for concurrency
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- 1. Tasks: The core unit of work
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    intent TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('todo', 'in_progress', 'review', 'human_review', 'done')),
    position INTEGER DEFAULT 0,
    agent_output TEXT,
    git_diff TEXT,
    message_history TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 2. Agent Logs: "The Matrix" Code Stream
CREATE TABLE IF NOT EXISTS agent_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level TEXT CHECK(level IN ('info', 'plan', 'code', 'error', 'success')),
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 3. GitHub Connections
CREATE TABLE IF NOT EXISTS github_connections (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('org', 'user')),
    login TEXT NOT NULL,
    avatar_url TEXT,
    token TEXT NOT NULL,
    scope TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 4. Projects (connected GitHub repos)
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    github_owner TEXT NOT NULL,
    github_repo TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_agent_logs_task ON agent_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_github_connections_type ON github_connections(type);
CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(github_owner);

-- 5. User Settings (BYOK)
CREATE TABLE IF NOT EXISTS user_settings (
    user_id TEXT PRIMARY KEY DEFAULT 'default',
    openrouter_key TEXT,
    zai_key TEXT,
    anthropic_key TEXT,
    openai_key TEXT,
    agent_backend TEXT DEFAULT 'native',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
