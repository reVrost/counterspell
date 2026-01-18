-- Revert tasks table changes

-- Remove current_step
ALTER TABLE tasks DROP COLUMN IF EXISTS current_step;

-- Restore old columns
ALTER TABLE tasks 
    ADD COLUMN IF NOT EXISTS agent_output TEXT,
    ADD COLUMN IF NOT EXISTS git_diff TEXT,
    ADD COLUMN IF NOT EXISTS message_history TEXT;

-- Revert status constraint
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check 
    CHECK(status IN ('planning', 'in_progress', 'agent_review', 'human_review', 'done'));

-- Migrate status values back
UPDATE tasks SET status = 'planning' WHERE status = 'pending';
UPDATE tasks SET status = 'human_review' WHERE status = 'review';
UPDATE tasks SET status = 'done' WHERE status = 'failed';
