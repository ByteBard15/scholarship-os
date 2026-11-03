ALTER TABLE writing_samples DROP CONSTRAINT writing_samples_document_type_check;
ALTER TABLE writing_samples ADD CONSTRAINT writing_samples_document_type_check CHECK (document_type IN ('personal_statement', 'motivation_letter', 'recommendation_letter', 'statement_of_purpose', 'essay', 'cover_letter', 'email', 'other'));
ALTER TABLE writing_samples ADD COLUMN source_profile_snapshot_id UUID NULL REFERENCES profile_snapshots(id) ON DELETE RESTRICT;
CREATE INDEX idx_writing_samples_source_profile_snapshot_id ON writing_samples(source_profile_snapshot_id);
