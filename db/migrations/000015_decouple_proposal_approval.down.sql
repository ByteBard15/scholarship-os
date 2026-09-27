DROP INDEX IF EXISTS idx_application_proposals_agent_queue;
DROP INDEX IF EXISTS idx_application_proposals_application_id;
DROP INDEX IF EXISTS idx_application_proposals_parent_profile_id;

ALTER TABLE application_proposals
    DROP COLUMN IF EXISTS application_id,
    DROP COLUMN IF EXISTS parent_profile_id;
