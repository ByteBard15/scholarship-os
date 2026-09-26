CREATE TABLE profile_overrides (
    id UUID PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id UUID,
    field_name TEXT NOT NULL,
    override_type TEXT NOT NULL,
    value JSONB,
    reason TEXT,
    source TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT profile_overrides_type_valid CHECK (override_type IN ('replace', 'hide', 'append')),
    CONSTRAINT profile_overrides_target_valid CHECK (override_type <> 'hide' OR entity_id IS NOT NULL)
);
CREATE INDEX profile_overrides_profile_id_idx ON profile_overrides(profile_id) WHERE deleted_at IS NULL;
CREATE INDEX profile_overrides_entity_idx ON profile_overrides(entity_type, entity_id) WHERE deleted_at IS NULL;

CREATE TABLE profile_snapshots (
    id UUID PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    reason TEXT,
    CONSTRAINT profile_snapshots_version_positive CHECK (version > 0),
    CONSTRAINT profile_snapshots_profile_version_unique UNIQUE (profile_id, version)
);
CREATE INDEX profile_snapshots_profile_id_idx ON profile_snapshots(profile_id, version DESC);
CREATE FUNCTION reject_profile_snapshot_update() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'profile snapshots are immutable';
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER profile_snapshots_immutable
    BEFORE UPDATE ON profile_snapshots
    FOR EACH ROW EXECUTE FUNCTION reject_profile_snapshot_update();

CREATE TABLE profile_documents (
    id UUID PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    document_type TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    storage_provider TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    sha256 TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profile_documents_type_valid CHECK (document_type IN ('cv', 'transcript', 'certificate', 'publication', 'portfolio', 'other')),
    CONSTRAINT profile_documents_status_valid CHECK (status IN ('uploaded', 'processing', 'ready', 'failed')),
    CONSTRAINT profile_documents_file_size_positive CHECK (file_size > 0),
    CONSTRAINT profile_documents_storage_key_unique UNIQUE (storage_key)
);
CREATE INDEX profile_documents_profile_type_idx ON profile_documents(profile_id, document_type);

CREATE TABLE profile_imports (
    id UUID PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES profile_documents(id) ON DELETE RESTRICT,
    import_type TEXT NOT NULL,
    status TEXT NOT NULL,
    raw_text TEXT,
    extracted_data JSONB,
    extraction_provider TEXT,
    extraction_model TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    CONSTRAINT profile_imports_status_valid CHECK (status IN ('uploaded', 'extracting', 'review_required', 'approved', 'partially_approved', 'rejected', 'failed')),
    CONSTRAINT profile_imports_type_valid CHECK (import_type IN ('cv'))
);
CREATE INDEX profile_imports_profile_status_idx ON profile_imports(profile_id, status);
CREATE INDEX profile_imports_document_id_idx ON profile_imports(document_id);

CREATE TABLE profile_import_candidates (
    id UUID PRIMARY KEY,
    import_id UUID NOT NULL REFERENCES profile_imports(id) ON DELETE CASCADE,
    section_type TEXT NOT NULL,
    candidate_data JSONB NOT NULL,
    source_text TEXT,
    confidence DOUBLE PRECISION,
    status TEXT NOT NULL,
    matched_entity_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profile_import_candidates_status_valid CHECK (status IN ('pending', 'accepted', 'rejected', 'merged')),
    CONSTRAINT profile_import_candidates_confidence_valid CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1))
);
CREATE INDEX profile_import_candidates_import_section_idx ON profile_import_candidates(import_id, section_type);
CREATE INDEX profile_import_candidates_import_status_idx ON profile_import_candidates(import_id, status);

CREATE TABLE profile_evidence (
    id UUID PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    field_name TEXT,
    evidence_type TEXT NOT NULL,
    document_id UUID REFERENCES profile_documents(id) ON DELETE SET NULL,
    import_id UUID REFERENCES profile_imports(id) ON DELETE SET NULL,
    source_url TEXT,
    source_text TEXT,
    confidence DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profile_evidence_confidence_valid CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1))
);
CREATE INDEX profile_evidence_profile_id_idx ON profile_evidence(profile_id);
CREATE INDEX profile_evidence_entity_idx ON profile_evidence(entity_type, entity_id);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    profile_id UUID REFERENCES applicant_profiles(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id UUID,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX audit_logs_profile_id_idx ON audit_logs(profile_id, created_at DESC) WHERE profile_id IS NOT NULL;
CREATE INDEX audit_logs_user_id_idx ON audit_logs(user_id, created_at DESC) WHERE user_id IS NOT NULL;
