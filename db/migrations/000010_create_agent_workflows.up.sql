CREATE TABLE research_tasks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_task_id UUID REFERENCES research_tasks(id) ON DELETE SET NULL,
    target_application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    instructions TEXT,
    task_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    priority TEXT,
    due_at TIMESTAMPTZ,
    research_config JSONB,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT research_tasks_status_check CHECK (status IN ('draft','ready','queued','running','review_required','completed','failed','cancelled')),
    CONSTRAINT research_tasks_parent_check CHECK (parent_task_id IS NULL OR parent_task_id <> id)
);
CREATE INDEX idx_research_tasks_user_id ON research_tasks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_research_tasks_status ON research_tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_research_tasks_task_type ON research_tasks(task_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_research_tasks_target_application ON research_tasks(target_application_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_research_tasks_parent ON research_tasks(parent_task_id) WHERE deleted_at IS NULL;

CREATE TABLE research_task_links (
    id UUID PRIMARY KEY,
    research_task_id UUID NOT NULL REFERENCES research_tasks(id) ON DELETE CASCADE,
    label TEXT,
    url TEXT NOT NULL,
    link_type TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_research_task_links_task ON research_task_links(research_task_id);

CREATE TABLE research_task_outputs (
    id UUID PRIMARY KEY,
    research_task_id UUID NOT NULL REFERENCES research_tasks(id) ON DELETE CASCADE,
    output_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_research_task_output UNIQUE(research_task_id, output_type, entity_id)
);
CREATE INDEX idx_research_task_outputs_task ON research_task_outputs(research_task_id);

ALTER TABLE research_runs ALTER COLUMN application_id DROP NOT NULL;
ALTER TABLE research_runs ADD COLUMN research_task_id UUID REFERENCES research_tasks(id) ON DELETE SET NULL;
CREATE INDEX idx_research_runs_task ON research_runs(research_task_id);
ALTER TABLE research_runs ADD CONSTRAINT research_runs_owner_check CHECK (application_id IS NOT NULL OR research_task_id IS NOT NULL);
ALTER TABLE research_sources ALTER COLUMN application_id DROP NOT NULL;
ALTER TABLE research_findings ALTER COLUMN application_id DROP NOT NULL;

CREATE TABLE application_proposals (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    research_task_id UUID NOT NULL REFERENCES research_tasks(id) ON DELETE CASCADE,
    research_run_id UUID REFERENCES research_runs(id) ON DELETE SET NULL,
    institution_id UUID REFERENCES institutions(id) ON DELETE SET NULL,
    programme_id UUID REFERENCES programmes(id) ON DELETE SET NULL,
    scholarship_id UUID REFERENCES scholarships(id) ON DELETE SET NULL,
    proposed_institution JSONB,
    proposed_programme JSONB,
    proposed_scholarship JSONB,
    name TEXT NOT NULL,
    country TEXT,
    intake TEXT,
    intake_year INTEGER,
    summary TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    confidence DOUBLE PRECISION,
    reasoning_summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    CONSTRAINT application_proposals_status_check CHECK (status IN ('pending','approved','rejected','superseded')),
    CONSTRAINT application_proposals_confidence_check CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1))
);
CREATE INDEX idx_application_proposals_user ON application_proposals(user_id);
CREATE INDEX idx_application_proposals_task ON application_proposals(research_task_id);
CREATE INDEX idx_application_proposals_status ON application_proposals(status);

CREATE TABLE application_proposal_sources (
    id UUID PRIMARY KEY,
    application_proposal_id UUID NOT NULL REFERENCES application_proposals(id) ON DELETE CASCADE,
    research_source_id UUID NOT NULL REFERENCES research_sources(id) ON DELETE CASCADE,
    source_role TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_application_proposal_source UNIQUE(application_proposal_id, research_source_id)
);

CREATE TABLE application_fields (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    value JSONB,
    value_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'empty',
    source_type TEXT,
    source_entity_id UUID,
    prefill_run_id UUID,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_application_field_key UNIQUE(application_id, key),
    CONSTRAINT application_fields_status_check CHECK (status IN ('empty','suggested','filled','verified','needs_review'))
);
CREATE INDEX idx_application_fields_application ON application_fields(application_id);
CREATE INDEX idx_application_fields_status ON application_fields(status);

CREATE TABLE application_questionnaires (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    questionnaire_type TEXT,
    status TEXT NOT NULL DEFAULT 'not_started',
    source_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT application_questionnaires_status_check CHECK (status IN ('not_started','in_progress','completed','needs_review'))
);
CREATE INDEX idx_application_questionnaires_application ON application_questionnaires(application_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_questionnaires_status ON application_questionnaires(status) WHERE deleted_at IS NULL;

CREATE TABLE application_questions (
    id UUID PRIMARY KEY,
    questionnaire_id UUID NOT NULL REFERENCES application_questionnaires(id) ON DELETE CASCADE,
    key TEXT,
    prompt TEXT NOT NULL,
    help_text TEXT,
    question_type TEXT NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    word_limit INTEGER,
    character_limit INTEGER,
    sort_order INTEGER,
    options JSONB,
    status TEXT NOT NULL DEFAULT 'unanswered',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT application_questions_status_check CHECK (status IN ('unanswered','suggested','answered','needs_information','needs_review','approved'))
);
CREATE INDEX idx_application_questions_questionnaire ON application_questions(questionnaire_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_application_questions_status ON application_questions(status) WHERE deleted_at IS NULL;

CREATE TABLE application_prefill_runs (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    research_task_id UUID REFERENCES research_tasks(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    trigger TEXT NOT NULL,
    agent_provider TEXT,
    agent_model TEXT,
    prompt_version TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT application_prefill_runs_status_check CHECK (status IN ('queued','running','review_required','complete','failed'))
);
CREATE INDEX idx_application_prefill_runs_application ON application_prefill_runs(application_id);
CREATE INDEX idx_application_prefill_runs_status ON application_prefill_runs(status);
CREATE INDEX idx_application_prefill_runs_created ON application_prefill_runs(created_at DESC);
ALTER TABLE application_fields ADD CONSTRAINT fk_application_fields_prefill_run FOREIGN KEY (prefill_run_id) REFERENCES application_prefill_runs(id) ON DELETE SET NULL;

CREATE TABLE application_answers (
    id UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES application_questions(id) ON DELETE CASCADE,
    value JSONB,
    draft_text TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    answer_source TEXT,
    source_entity_type TEXT,
    source_entity_id UUID,
    confidence DOUBLE PRECISION,
    created_by TEXT,
    prefill_run_id UUID REFERENCES application_prefill_runs(id) ON DELETE SET NULL,
    agent_provider TEXT,
    agent_model TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT application_answers_status_check CHECK (status IN ('draft','suggested','answered','approved','rejected','superseded')),
    CONSTRAINT application_answers_confidence_check CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1))
);
CREATE INDEX idx_application_answers_question ON application_answers(question_id);
CREATE INDEX idx_application_answers_status ON application_answers(status);

CREATE TABLE information_requests (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    research_task_id UUID REFERENCES research_tasks(id) ON DELETE SET NULL,
    questionnaire_id UUID REFERENCES application_questionnaires(id) ON DELETE CASCADE,
    question_id UUID REFERENCES application_questions(id) ON DELETE CASCADE,
    application_field_id UUID REFERENCES application_fields(id) ON DELETE CASCADE,
    request_type TEXT NOT NULL,
    title TEXT NOT NULL,
    prompt TEXT NOT NULL,
    context TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    priority TEXT,
    response_type TEXT NOT NULL,
    options JSONB,
    response_value JSONB,
    response_text TEXT,
    created_by TEXT NOT NULL,
    resolution_source TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    reopened_at TIMESTAMPTZ,
    CONSTRAINT information_requests_status_check CHECK (status IN ('pending','answered','processing','completed','reopened','cancelled'))
);
CREATE INDEX idx_information_requests_user ON information_requests(user_id);
CREATE INDEX idx_information_requests_application ON information_requests(application_id);
CREATE INDEX idx_information_requests_status ON information_requests(status);
CREATE INDEX idx_information_requests_question ON information_requests(question_id);
CREATE INDEX idx_information_requests_field ON information_requests(application_field_id);
CREATE UNIQUE INDEX uq_information_requests_open_question ON information_requests(user_id, application_id, question_id, request_type) WHERE question_id IS NOT NULL AND status IN ('pending','answered','processing','reopened');
CREATE UNIQUE INDEX uq_information_requests_open_field ON information_requests(user_id, application_id, application_field_id, request_type) WHERE application_field_id IS NOT NULL AND status IN ('pending','answered','processing','reopened');

CREATE TABLE information_request_responses (
    id UUID PRIMARY KEY,
    information_request_id UUID NOT NULL REFERENCES information_requests(id) ON DELETE CASCADE,
    response_value JSONB,
    response_text TEXT,
    submitted_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_information_request_responses_request ON information_request_responses(information_request_id, created_at);

CREATE TABLE prefill_actions (
    id UUID PRIMARY KEY,
    prefill_run_id UUID NOT NULL REFERENCES application_prefill_runs(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id UUID,
    status TEXT NOT NULL,
    summary TEXT,
    source_type TEXT,
    source_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_prefill_actions_run ON prefill_actions(prefill_run_id);

CREATE TABLE agent_activities (
    id UUID PRIMARY KEY,
    research_task_id UUID REFERENCES research_tasks(id) ON DELETE SET NULL,
    application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    prefill_run_id UUID REFERENCES application_prefill_runs(id) ON DELETE SET NULL,
    activity_type TEXT NOT NULL,
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_agent_activities_task ON agent_activities(research_task_id, created_at);
CREATE INDEX idx_agent_activities_application ON agent_activities(application_id, created_at);

