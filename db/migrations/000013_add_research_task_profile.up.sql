ALTER TABLE research_tasks
    ADD COLUMN profile_id UUID NULL REFERENCES applicant_profiles(id) ON DELETE SET NULL;

CREATE INDEX idx_research_tasks_profile ON research_tasks(profile_id);
