CREATE TABLE research_runs (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued',
    trigger TEXT NOT NULL DEFAULT 'manual',
    research_type TEXT NOT NULL DEFAULT 'full',
    model_provider TEXT,
    model_name TEXT,
    prompt_version TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT research_runs_status_check CHECK (status IN ('queued','running','review_required','complete','failed'))
);
CREATE INDEX idx_research_runs_application ON research_runs(application_id);
CREATE INDEX idx_research_runs_status ON research_runs(status);
CREATE INDEX idx_research_runs_created_at ON research_runs(created_at DESC);

CREATE TABLE research_sources (
    id UUID PRIMARY KEY,
    research_run_id UUID NOT NULL REFERENCES research_runs(id) ON DELETE CASCADE,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title TEXT,
    publisher TEXT,
    source_type TEXT NOT NULL,
    is_official BOOLEAN NOT NULL DEFAULT FALSE,
    retrieved_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ,
    content_hash TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_research_sources_run ON research_sources(research_run_id);
CREATE INDEX idx_research_sources_application ON research_sources(application_id);

CREATE TABLE research_findings (
    id UUID PRIMARY KEY,
    research_run_id UUID NOT NULL REFERENCES research_runs(id) ON DELETE CASCADE,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    source_id UUID REFERENCES research_sources(id) ON DELETE SET NULL,
    category TEXT NOT NULL,
    field TEXT NOT NULL,
    value JSONB NOT NULL,
    confidence DOUBLE PRECISION CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1)),
    verification_status TEXT NOT NULL DEFAULT 'unverified',
    review_status TEXT NOT NULL DEFAULT 'pending',
    raw_text TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT research_findings_verification_check CHECK (verification_status IN ('verified','supported','conflicting','unverified','inferred')),
    CONSTRAINT research_findings_review_check CHECK (review_status IN ('pending','accepted','rejected','auto_accepted'))
);
CREATE INDEX idx_research_findings_run ON research_findings(research_run_id);
CREATE INDEX idx_research_findings_application ON research_findings(application_id);
CREATE INDEX idx_research_findings_category ON research_findings(category);
CREATE INDEX idx_research_findings_review ON research_findings(review_status);

ALTER TABLE application_requirements ADD CONSTRAINT fk_application_requirements_source FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE SET NULL;
ALTER TABLE application_deadlines ADD CONSTRAINT fk_application_deadlines_source FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE SET NULL;
ALTER TABLE application_funding ADD CONSTRAINT fk_application_funding_source FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE SET NULL;
ALTER TABLE application_contacts ADD CONSTRAINT fk_application_contacts_source FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE SET NULL;
ALTER TABLE application_supervisors ADD CONSTRAINT fk_application_supervisors_source FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE SET NULL;
