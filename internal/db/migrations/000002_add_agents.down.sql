-- Rollback agents table and task assignment columns

DROP INDEX IF EXISTS idx_tasks_assigned_user;
DROP INDEX IF EXISTS idx_tasks_assigned_agent;
DROP INDEX IF EXISTS idx_agents_user;

ALTER TABLE tasks 
    DROP COLUMN IF EXISTS assigned_user_id,
    DROP COLUMN IF EXISTS assigned_agent_id;

DROP TABLE IF EXISTS agents;
