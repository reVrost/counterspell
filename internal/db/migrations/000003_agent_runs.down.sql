-- Rollback: restore agent_logs

DROP TABLE IF EXISTS agent_runs;

CREATE TABLE IF NOT EXISTS agent_logs (
    id SERIAL PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level TEXT CHECK(level IN ('info', 'plan', 'code', 'error', 'success')),
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_logs_task ON agent_logs(task_id);
