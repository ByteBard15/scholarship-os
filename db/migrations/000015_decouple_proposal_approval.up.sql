ALTER TABLE application_proposals
    ADD COLUMN parent_profile_id UUID NULL REFERENCES applicant_profiles(id) ON DELETE RESTRICT,
    ADD COLUMN application_id UUID NULL REFERENCES applications(id) ON DELETE SET NULL;

CREATE INDEX idx_application_proposals_parent_profile_id ON application_proposals(parent_profile_id);

-- Before this migration, approval created an Application immediately. Mark
-- those legacy proposals as materialized using their retained task output so
-- an agent cannot create a duplicate after deployment.
UPDATE application_proposals AS proposal
SET application_id = output.entity_id
FROM research_task_outputs AS output
WHERE proposal.status = 'approved'
  AND proposal.application_id IS NULL
  AND output.research_task_id = proposal.research_task_id
  AND output.output_type = 'application';

UPDATE application_proposals AS proposal
SET parent_profile_id = application_profile.parent_profile_id
FROM applications AS application,
     applicant_profiles AS application_profile
WHERE proposal.application_id = application.id
  AND application.applicant_profile_id = application_profile.id
  AND proposal.parent_profile_id IS NULL;

CREATE INDEX idx_application_proposals_application_id ON application_proposals(application_id);
CREATE INDEX idx_application_proposals_agent_queue
    ON application_proposals(user_id, status, created_at)
    WHERE status = 'approved' AND application_id IS NULL;
