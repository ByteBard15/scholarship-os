CREATE TABLE applications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    applicant_profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE RESTRICT,
    institution_id UUID REFERENCES institutions(id) ON DELETE SET NULL,
    programme_id UUID REFERENCES programmes(id) ON DELETE SET NULL,
    scholarship_id UUID REFERENCES scholarships(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    intake TEXT,
    intake_year INTEGER,
    country TEXT,
    status TEXT NOT NULL DEFAULT 'discovered',
    priority INTEGER,
    notes TEXT,
    research_status TEXT NOT NULL DEFAULT 'not_started',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT applications_status_check CHECK (status IN ('discovered','researching','research_complete','eligibility_review','preparing','documents_in_progress','ready_to_submit','submitted','waiting','accepted','rejected','withdrawn','expired')),
    CONSTRAINT applications_research_status_check CHECK (research_status IN ('not_started','queued','running','review_required','complete','failed','stale')),
    CONSTRAINT applications_priority_check CHECK (priority IS NULL OR priority BETWEEN 1 AND 5)
);

CREATE INDEX idx_applications_user_id ON applications(user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_applications_profile_id ON applications(applicant_profile_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_institution_id ON applications(institution_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_programme_id ON applications(programme_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_scholarship_id ON applications(scholarship_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_status ON applications(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_research_status ON applications(research_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_intake_year ON applications(intake_year) WHERE deleted_at IS NULL;

CREATE TABLE application_requirements (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    status TEXT NOT NULL DEFAULT 'unknown',
    due_date DATE,
    source_id UUID,
    source_url TEXT,
    evidence_required BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT,
    sort_order INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT application_requirements_status_check CHECK (status IN ('unknown','not_started','in_progress','satisfied','not_satisfied','not_applicable','waived'))
);
CREATE INDEX idx_application_requirements_application ON application_requirements(application_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_requirements_status ON application_requirements(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_requirements_category ON application_requirements(category) WHERE deleted_at IS NULL;

CREATE TABLE application_deadlines (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    deadline_type TEXT NOT NULL,
    title TEXT NOT NULL,
    deadline_at TIMESTAMPTZ,
    timezone TEXT,
    date_precision TEXT NOT NULL DEFAULT 'unknown',
    raw_deadline_text TEXT,
    is_hard_deadline BOOLEAN NOT NULL DEFAULT FALSE,
    source_id UUID,
    source_url TEXT,
    verified_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT application_deadlines_precision_check CHECK (date_precision IN ('exact','day','month','approximate','unknown'))
);
CREATE INDEX idx_application_deadlines_application ON application_deadlines(application_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_deadlines_at ON application_deadlines(deadline_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_deadlines_type ON application_deadlines(deadline_type) WHERE deleted_at IS NULL;

CREATE TABLE application_funding (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    funding_type TEXT NOT NULL,
    currency TEXT,
    amount NUMERIC(14,2),
    amount_period TEXT,
    tuition_coverage TEXT,
    stipend_amount NUMERIC(14,2),
    stipend_period TEXT,
    travel_coverage TEXT,
    insurance_coverage TEXT,
    accommodation_coverage TEXT,
    other_benefits TEXT,
    conditions TEXT,
    source_id UUID,
    source_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_application_funding_application ON application_funding(application_id) WHERE deleted_at IS NULL;

CREATE TABLE application_contacts (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name TEXT,
    role TEXT,
    email TEXT,
    phone TEXT,
    organization TEXT,
    contact_type TEXT,
    url TEXT,
    notes TEXT,
    source_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_application_contacts_application ON application_contacts(application_id) WHERE deleted_at IS NULL;

CREATE TABLE application_supervisors (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    title TEXT,
    department TEXT,
    institution TEXT,
    email TEXT,
    profile_url TEXT,
    research_areas TEXT,
    contact_status TEXT,
    notes TEXT,
    source_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_application_supervisors_application ON application_supervisors(application_id) WHERE deleted_at IS NULL;

CREATE TABLE application_urls (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    url_type TEXT NOT NULL,
    label TEXT,
    url TEXT NOT NULL,
    is_official BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_application_urls_application ON application_urls(application_id) WHERE deleted_at IS NULL;

CREATE TABLE application_tasks (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    parent_task_id UUID REFERENCES application_tasks(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    priority INTEGER,
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    sort_order INTEGER,
    task_type TEXT,
    source TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT application_tasks_status_check CHECK (status IN ('todo','in_progress','blocked','done','cancelled')),
    CONSTRAINT application_tasks_priority_check CHECK (priority IS NULL OR priority BETWEEN 1 AND 5)
);
CREATE INDEX idx_application_tasks_application ON application_tasks(application_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_tasks_status ON application_tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_tasks_due_at ON application_tasks(due_at) WHERE deleted_at IS NULL;

CREATE TABLE external_task_references (
    id UUID PRIMARY KEY,
    application_task_id UUID NOT NULL REFERENCES application_tasks(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    external_list_id TEXT,
    external_task_id TEXT NOT NULL,
    external_url TEXT,
    last_synced_at TIMESTAMPTZ,
    sync_status TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_external_task_provider UNIQUE(provider, external_task_id)
);
CREATE INDEX idx_external_task_references_task ON external_task_references(application_task_id);

CREATE TABLE requirement_evidence (
    id UUID PRIMARY KEY,
    requirement_id UUID NOT NULL REFERENCES application_requirements(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    evidence_id UUID REFERENCES profile_evidence(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'suggested',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT requirement_evidence_status_check CHECK (status IN ('suggested','accepted','rejected'))
);
CREATE INDEX idx_requirement_evidence_requirement ON requirement_evidence(requirement_id);
CREATE INDEX idx_requirement_evidence_profile ON requirement_evidence(profile_id);
