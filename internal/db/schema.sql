-- SQLite Schema for Counterspell (Local-First Data Plane)
-- This is the consolidated schema - no migrations needed

-- Machines: Track local and cloud worker instances
CREATE TABLE IF NOT EXISTS machines (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,  -- e.g., "Alice's MacBook Pro", "Fly Agent #1"
    mode TEXT CHECK(mode IN ('local', 'cloud')) NOT NULL,
    capabilities TEXT,  -- JSON: {"os": "darwin", "cpus": 8}
    last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Agents: System-wide agent configurations (no user_id - single-tenant)
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    system_prompt TEXT NOT NULL,
    tools TEXT NOT NULL DEFAULT '{}',  -- JSON array
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tasks: Core unit of work
-- Status flow: pending -> in_progress -> review -> done | failed
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    machine_id TEXT NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    intent TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'in_progress', 'review', 'done', 'failed')),
    position INTEGER DEFAULT 0,
    current_step TEXT,
    assigned_agent_id TEXT REFERENCES agents(id) ON DELETE SET NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Agent Runs: One row per agent execution within a task
CREATE TABLE IF NOT EXISTS agent_runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    step TEXT NOT NULL,
    agent_id TEXT REFERENCES agents(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'completed', 'failed')),
    input TEXT,
    output TEXT,
    message_history TEXT,  -- JSON
    artifact_path TEXT,
    error TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Settings: API keys and configuration (no user_id - single-tenant)
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    openrouter_key TEXT,
    zai_key TEXT,
    anthropic_key TEXT,
    openai_key TEXT,
    agent_backend TEXT DEFAULT 'native',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings row
INSERT OR IGNORE INTO settings (id, agent_backend) VALUES (1, 'native');

-- Indices (optimize for common query patterns)
CREATE INDEX IF NOT EXISTS idx_tasks_machine_status ON tasks(machine_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_machine ON tasks(machine_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_agent ON tasks(assigned_agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_runs_task ON agent_runs(task_id);
CREATE INDEX IF NOT EXISTS idx_agent_runs_task_step ON agent_runs(task_id, step);
CREATE INDEX IF NOT EXISTS idx_agent_runs_status ON agent_runs(status);
CREATE INDEX IF NOT EXISTS idx_machines_mode ON machines(mode);
