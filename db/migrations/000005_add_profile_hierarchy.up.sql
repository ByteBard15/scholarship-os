ALTER TABLE applicant_profiles
    ADD COLUMN profile_type TEXT NOT NULL DEFAULT 'master',
    ADD COLUMN parent_profile_id UUID REFERENCES applicant_profiles(id) ON DELETE RESTRICT;

ALTER TABLE applicant_profiles
    ADD CONSTRAINT applicant_profiles_type_valid
        CHECK (profile_type IN ('master', 'domain', 'application')),
    ADD CONSTRAINT applicant_profiles_parent_rules
        CHECK (
            (profile_type = 'master' AND parent_profile_id IS NULL)
            OR (profile_type IN ('domain', 'application') AND parent_profile_id IS NOT NULL)
        ),
    ADD CONSTRAINT applicant_profiles_not_self_parent
        CHECK (parent_profile_id IS NULL OR parent_profile_id <> id);

CREATE INDEX applicant_profiles_profile_type_idx
    ON applicant_profiles(profile_type) WHERE deleted_at IS NULL;
CREATE INDEX applicant_profiles_parent_profile_id_idx
    ON applicant_profiles(parent_profile_id) WHERE parent_profile_id IS NOT NULL AND deleted_at IS NULL;
