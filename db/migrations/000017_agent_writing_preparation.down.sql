UPDATE writing_samples SET document_type = 'other' WHERE document_type = 'email';
ALTER TABLE writing_samples DROP COLUMN source_profile_snapshot_id;
ALTER TABLE writing_samples DROP CONSTRAINT writing_samples_document_type_check;
ALTER TABLE writing_samples ADD CONSTRAINT writing_samples_document_type_check CHECK (document_type IN ('personal_statement', 'motivation_letter', 'recommendation_letter', 'statement_of_purpose', 'essay', 'cover_letter', 'other'));
