-- SQLite Schema for Counterspell (Local-First Data Plane)
-- This is consolidated schema - no migrations needed

-- Tasks: Core unit of work
-- Status flow: planning -> in_progress -> review -> done | failed
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    intent TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('planning', 'in_progress', 'review', 'done', 'failed')),
    position INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL, -- timestampz replacement is unix in milli,
    updated_at INTEGER NOT NULL -- timestampz replacement is unix in milli
);

CREATE TRIGGER IF NOT EXISTS update_tasks_updated_at
AFTER UPDATE ON tasks
BEGIN
UPDATE tasks SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

-- Agent Runs: One row per agent execution within a task
CREATE TABLE IF NOT EXISTS agent_runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    agent_backend TEXT NOT NULL CHECK(agent_backend IN ('native', 'claude-code', 'codex')),
    summary_message_id TEXT REFERENCES messages(id) ON DELETE SET NULL,
    cost REAL NOT NULL DEFAULT 0.0 CHECK (cost >= 0.0),
    message_count INTEGER NOT NULL DEFAULT 0 CHECK (message_count >= 0),
    prompt_tokens  INTEGER NOT NULL DEFAULT 0 CHECK (prompt_tokens >= 0),
    completion_tokens  INTEGER NOT NULL DEFAULT 0 CHECK (completion_tokens>= 0),
    completed_at DATETIME,
    created_at INTEGER NOT NULL, -- timestampz replacement is unix in milli,
    updated_at INTEGER NOT NULL -- timestampz replacement is unix in milli
);

CREATE TRIGGER IF NOT EXISTS update_agent_runs_updated_at
AFTER UPDATE ON agent_runs
BEGIN
UPDATE agent_runs SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

-- Artifacts: Files uploaded by agents
CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    updated_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    UNIQUE(path, run_id, version)
);

CREATE TRIGGER IF NOT EXISTS update_artifacts_updated_at
AFTER UPDATE ON artifacts
BEGIN
UPDATE artifacts SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

-- Messages: Chat messages for agent conversation history
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    run_id TEXT REFERENCES agent_runs(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('system', 'user', 'assistant', 'tool')),
    parts TEXT NOT NULL default '[]',
    model TEXT,
    provider TEXT,
    content TEXT NOT NULL,
    tool_id TEXT,  -- For tool messages: which tool was called
    created_at INTEGER NOT NULL, -- timestampz replacement is unix in milli,
    updated_at INTEGER NOT NULL, -- timestampz replacement is unix in milli,
    finished_at INTEGER
);

CREATE TRIGGER IF NOT EXISTS update_messages_updated_at
AFTER UPDATE ON messages
BEGIN
UPDATE messages SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

CREATE TRIGGER IF NOT EXISTS update_run_message_count_on_insert
AFTER INSERT ON messages
BEGIN
UPDATE agent_runs SET
    message_count = message_count + 1
WHERE id = new.run_id;
END;

CREATE TRIGGER IF NOT EXISTS update_run_message_count_on_delete
AFTER DELETE ON messages
BEGIN
UPDATE agent_runs SET
    message_count = message_count - 1
WHERE id = old.run_id;
END;

-- Settings: API keys and configuration (no user_id - single-tenant)
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    openrouter_key TEXT,
    zai_key TEXT,
    anthropic_key TEXT,
    openai_key TEXT,
    agent_backend TEXT NOT NULL CHECK(agent_backend IN ('native', 'claude-code', 'codex')),
    updated_at INTEGER NOT NULL -- timestampz replacement is unix in milli
);

CREATE TRIGGER IF NOT EXISTS update_settings_updated_at
AFTER UPDATE ON settings
BEGIN
UPDATE settings SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

-- Insert default settings row
INSERT OR IGNORE INTO settings (id, agent_backend) VALUES (1, 'native');

-- Indices (optimize for common query patterns)
CREATE INDEX IF NOT EXISTS idx_agent_runs_task ON agent_runs(task_id);
CREATE INDEX IF NOT EXISTS idx_messages_task ON messages(task_id);
CREATE INDEX IF NOT EXISTS idx_messages_task_created ON messages(task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_run ON messages(run_id);
CREATE INDEX IF NOT EXISTS idx_runs_created_at ON agent_runs (created_at);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);
CREATE INDEX IF NOT EXISTS idx_artifacts_created_at ON artifacts (created_at);
