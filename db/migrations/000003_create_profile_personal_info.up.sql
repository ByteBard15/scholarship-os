CREATE TABLE profile_personal_info (
    id UUID PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    last_name TEXT NOT NULL,
    preferred_name TEXT,
    phone TEXT,
    city TEXT,
    state_or_region TEXT,
    country TEXT,
    nationality TEXT,
    linked_in_url TEXT,
    github_url TEXT,
    website_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profile_personal_info_profile_unique UNIQUE(profile_id)
);
