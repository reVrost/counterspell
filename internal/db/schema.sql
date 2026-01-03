-- Enforce WAL mode for concurrency
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- 1. Tasks: The core unit of work
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    intent TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('todo', 'in_progress', 'review', 'human_review', 'done')),
    position INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. Agent Logs: "The Matrix" Code Stream
CREATE TABLE IF NOT EXISTS agent_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level TEXT CHECK(level IN ('info', 'plan', 'code', 'error', 'success')),
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_agent_logs_task ON agent_logs(task_id);
