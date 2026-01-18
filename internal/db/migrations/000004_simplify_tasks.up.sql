-- Simplify tasks table: remove old agent columns, add current_step, update status values

-- Drop old columns no longer needed (message_history moved to agent_runs)
ALTER TABLE tasks 
    DROP COLUMN IF EXISTS agent_output,
    DROP COLUMN IF EXISTS git_diff,
    DROP COLUMN IF EXISTS message_history;

-- Add current_step for workflow tracking
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS current_step TEXT;

-- Update status constraint to new values
-- First drop existing constraint, then add new one
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check 
    CHECK(status IN ('pending', 'in_progress', 'review', 'done', 'failed'));

-- Migrate existing status values
UPDATE tasks SET status = 'pending' WHERE status = 'planning';
UPDATE tasks SET status = 'review' WHERE status IN ('agent_review', 'human_review');
