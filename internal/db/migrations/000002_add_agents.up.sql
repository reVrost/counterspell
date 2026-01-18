-- Add agents table and task assignment columns

-- Agents: System-wide agent configurations
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT 'system',
    name TEXT NOT NULL UNIQUE,
    system_prompt TEXT NOT NULL,
    tools TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_tools CHECK (
        tools <@ ARRAY['bash', 'edit', 'glob', 'grep', 'ls', 'multiedit', 'read', 'todo', 'write']::TEXT[]
    )
);

-- Add assignment columns to tasks
ALTER TABLE tasks 
    ADD COLUMN assigned_agent_id TEXT REFERENCES agents(id) ON DELETE SET NULL,
    ADD COLUMN assigned_user_id TEXT;

-- Index for agent lookups
CREATE INDEX IF NOT EXISTS idx_agents_user ON agents(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_agent ON tasks(assigned_agent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_user ON tasks(assigned_user_id);
