ALTER TABLE application_proposals
    ADD COLUMN priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    ADD COLUMN rank INTEGER NULL;

ALTER TABLE application_proposals
    ADD CONSTRAINT application_proposals_priority_check
        CHECK (priority IN ('highest', 'high', 'medium', 'low')),
    ADD CONSTRAINT application_proposals_rank_check
        CHECK (rank IS NULL OR rank >= 1);

CREATE INDEX idx_application_proposals_task_priority
    ON application_proposals(research_task_id, priority, rank);
