-- PostgreSQL Schema for Counterspell
-- All user-scoped tables include user_id for multi-tenancy

-- Projects (connected GitHub repos)
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    github_owner TEXT NOT NULL,
    github_repo TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, github_owner, github_repo)
);

-- Tasks: Core unit of work
-- Status flow: planning → in_progress → agent_review → human_review → done
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    intent TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('planning', 'in_progress', 'agent_review', 'human_review', 'done')),
    position INTEGER DEFAULT 0,
    agent_output TEXT,
    git_diff TEXT,
    message_history TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agent_logs (
    id SERIAL PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level TEXT CHECK(level IN ('info', 'plan', 'code', 'error', 'success')),
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- GitHub Connections (OAuth tokens)
CREATE TABLE IF NOT EXISTS github_connections (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('org', 'user')),
    login TEXT NOT NULL,
    avatar_url TEXT,
    token TEXT NOT NULL,
    scope TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Settings (BYOK - Bring Your Own Keys)
CREATE TABLE IF NOT EXISTS user_settings (
    user_id TEXT PRIMARY KEY,
    openrouter_key TEXT,
    zai_key TEXT,
    anthropic_key TEXT,
    openai_key TEXT,
    agent_backend TEXT DEFAULT 'native',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Repo Cache: Cached GitHub repo metadata
CREATE TABLE IF NOT EXISTS repo_cache (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    last_fetched_at TIMESTAMPTZ,
    is_favorite BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, owner, name)
);

-- Indices (always filter by user_id first)
CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks(user_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_user_project ON tasks(user_id, project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_agent_logs_task ON agent_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_github_connections_user ON github_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_repo_cache_user ON repo_cache(user_id);
CREATE INDEX IF NOT EXISTS idx_repo_cache_favorite ON repo_cache(user_id, is_favorite);
