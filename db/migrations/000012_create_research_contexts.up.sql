CREATE TABLE research_contexts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_research_contexts_user
    ON research_contexts(user_id)
    WHERE deleted_at IS NULL;

CREATE TABLE research_task_contexts (
    research_task_id UUID NOT NULL REFERENCES research_tasks(id) ON DELETE CASCADE,
    research_context_id UUID NOT NULL REFERENCES research_contexts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (research_task_id, research_context_id)
);

CREATE INDEX idx_research_task_contexts_context
    ON research_task_contexts(research_context_id);
