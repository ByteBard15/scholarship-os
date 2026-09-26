DROP INDEX IF EXISTS applicant_profiles_parent_profile_id_idx;
DROP INDEX IF EXISTS applicant_profiles_profile_type_idx;
ALTER TABLE applicant_profiles
    DROP CONSTRAINT IF EXISTS applicant_profiles_not_self_parent,
    DROP CONSTRAINT IF EXISTS applicant_profiles_parent_rules,
    DROP CONSTRAINT IF EXISTS applicant_profiles_type_valid,
    DROP COLUMN IF EXISTS parent_profile_id,
    DROP COLUMN IF EXISTS profile_type;
