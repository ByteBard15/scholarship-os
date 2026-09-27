DROP INDEX IF EXISTS idx_application_proposals_task_priority;

ALTER TABLE application_proposals
    DROP CONSTRAINT IF EXISTS application_proposals_rank_check,
    DROP CONSTRAINT IF EXISTS application_proposals_priority_check,
    DROP COLUMN IF EXISTS rank,
    DROP COLUMN IF EXISTS priority;
