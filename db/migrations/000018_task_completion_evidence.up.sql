ALTER TABLE application_tasks ADD COLUMN completion_criteria TEXT;
CREATE TABLE application_task_evidence (
    id UUID PRIMARY KEY,
    application_task_id UUID NOT NULL REFERENCES application_tasks(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('link', 'identifier', 'document', 'confirmation', 'note')),
    label TEXT NOT NULL,
    value TEXT NOT NULL,
    source_url TEXT,
    recorded_by UUID,
    recorded_by_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX application_task_evidence_task_idx ON application_task_evidence(application_task_id, created_at);
