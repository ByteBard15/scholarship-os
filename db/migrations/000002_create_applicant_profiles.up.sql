CREATE TABLE applicant_profiles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    headline TEXT,
    summary TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX applicant_profiles_user_id_idx ON applicant_profiles(user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX applicant_profiles_one_default_per_user_idx ON applicant_profiles(user_id) WHERE is_default AND deleted_at IS NULL;
