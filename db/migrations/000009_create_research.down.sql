ALTER TABLE application_supervisors DROP CONSTRAINT IF EXISTS fk_application_supervisors_source;
ALTER TABLE application_contacts DROP CONSTRAINT IF EXISTS fk_application_contacts_source;
ALTER TABLE application_funding DROP CONSTRAINT IF EXISTS fk_application_funding_source;
ALTER TABLE application_deadlines DROP CONSTRAINT IF EXISTS fk_application_deadlines_source;
ALTER TABLE application_requirements DROP CONSTRAINT IF EXISTS fk_application_requirements_source;
DROP TABLE IF EXISTS research_findings;
DROP TABLE IF EXISTS research_sources;
DROP TABLE IF EXISTS research_runs;
