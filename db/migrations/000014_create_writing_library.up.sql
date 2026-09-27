CREATE TABLE writing_samples (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id UUID NULL REFERENCES applications(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    origin VARCHAR(30) NOT NULL DEFAULT 'user',
    status VARCHAR(30) NOT NULL DEFAULT 'source',
    description TEXT NULL,
    created_by VARCHAR(100) NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT writing_samples_document_type_check CHECK (document_type IN ('personal_statement', 'motivation_letter', 'recommendation_letter', 'statement_of_purpose', 'essay', 'cover_letter', 'other')),
    CONSTRAINT writing_samples_origin_check CHECK (origin IN ('user', 'agent', 'import')),
    CONSTRAINT writing_samples_status_check CHECK (status IN ('source', 'draft', 'suggested', 'approved', 'archived'))
);

CREATE INDEX idx_writing_samples_user_id ON writing_samples(user_id);
CREATE INDEX idx_writing_samples_application_id ON writing_samples(application_id);
CREATE INDEX idx_writing_samples_document_type ON writing_samples(document_type);
CREATE INDEX idx_writing_samples_status ON writing_samples(status);

CREATE TABLE writing_tags (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    normalized_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_writing_tags_user_normalized UNIQUE (user_id, normalized_name)
);

CREATE INDEX idx_writing_tags_user_id ON writing_tags(user_id);

CREATE TABLE writing_sample_tags (
    writing_sample_id UUID NOT NULL REFERENCES writing_samples(id) ON DELETE CASCADE,
    writing_tag_id UUID NOT NULL REFERENCES writing_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (writing_sample_id, writing_tag_id)
);

CREATE INDEX idx_writing_sample_tags_tag_id ON writing_sample_tags(writing_tag_id);

CREATE TABLE writing_sample_relations (
    id UUID PRIMARY KEY,
    source_sample_id UUID NOT NULL REFERENCES writing_samples(id) ON DELETE CASCADE,
    derived_sample_id UUID NOT NULL REFERENCES writing_samples(id) ON DELETE CASCADE,
    relation_type VARCHAR(30) NOT NULL DEFAULT 'derived_from',
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT writing_sample_relations_not_self CHECK (source_sample_id <> derived_sample_id),
    CONSTRAINT uq_writing_sample_relation UNIQUE (source_sample_id, derived_sample_id, relation_type)
);

CREATE INDEX idx_writing_sample_relations_derived ON writing_sample_relations(derived_sample_id);
