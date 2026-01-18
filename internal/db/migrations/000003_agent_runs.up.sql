-- Replace agent_logs with agent_runs
-- One row per agent execution, message history as JSON

DROP TABLE IF EXISTS agent_logs;

CREATE TABLE agent_runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    step TEXT NOT NULL,
    agent_id TEXT REFERENCES agents(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'completed', 'failed')),
    input TEXT,
    output TEXT,
    message_history JSONB DEFAULT '[]',
    artifact_path TEXT,
    error TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agent_runs_task ON agent_runs(task_id);
CREATE INDEX idx_agent_runs_task_step ON agent_runs(task_id, step);
CREATE INDEX idx_agent_runs_status ON agent_runs(status);
