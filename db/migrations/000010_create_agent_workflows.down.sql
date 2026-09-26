DROP TABLE IF EXISTS agent_activities;
DROP TABLE IF EXISTS prefill_actions;
DROP TABLE IF EXISTS information_request_responses;
DROP TABLE IF EXISTS information_requests;
DROP TABLE IF EXISTS application_answers;
ALTER TABLE application_fields DROP CONSTRAINT IF EXISTS fk_application_fields_prefill_run;
DROP TABLE IF EXISTS application_prefill_runs;
DROP TABLE IF EXISTS application_questions;
DROP TABLE IF EXISTS application_questionnaires;
DROP TABLE IF EXISTS application_fields;
DROP TABLE IF EXISTS application_proposal_sources;
DROP TABLE IF EXISTS application_proposals;

DELETE FROM research_runs WHERE application_id IS NULL;
ALTER TABLE research_runs DROP CONSTRAINT IF EXISTS research_runs_owner_check;
DROP INDEX IF EXISTS idx_research_runs_task;
ALTER TABLE research_runs DROP COLUMN IF EXISTS research_task_id;
ALTER TABLE research_findings ALTER COLUMN application_id SET NOT NULL;
ALTER TABLE research_sources ALTER COLUMN application_id SET NOT NULL;
ALTER TABLE research_runs ALTER COLUMN application_id SET NOT NULL;

DROP TABLE IF EXISTS research_task_outputs;
DROP TABLE IF EXISTS research_task_links;
DROP TABLE IF EXISTS research_tasks;
