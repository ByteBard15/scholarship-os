CREATE TABLE education_history (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 institution TEXT NOT NULL, degree TEXT NOT NULL, field_of_study TEXT NOT NULL,
 start_date DATE, end_date DATE, is_current BOOLEAN NOT NULL DEFAULT FALSE,
 grade TEXT, grade_scale TEXT, classification TEXT, city TEXT, country TEXT, description TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT education_dates_valid CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX education_history_profile_id_idx ON education_history(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE employment_history (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 organization TEXT NOT NULL, job_title TEXT NOT NULL, employment_type TEXT,
 start_date DATE, end_date DATE, is_current BOOLEAN NOT NULL DEFAULT FALSE, city TEXT, country TEXT, description TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT employment_dates_valid CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX employment_history_profile_id_idx ON employment_history(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE projects (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 title TEXT NOT NULL, organization TEXT, start_date DATE, end_date DATE, description TEXT, url TEXT, repository_url TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT project_dates_valid CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX projects_profile_id_idx ON projects(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE publications (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 title TEXT NOT NULL, publication_type TEXT, publisher TEXT, publication_date DATE, doi TEXT, url TEXT, description TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ
);
CREATE INDEX publications_profile_id_idx ON publications(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE articles (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 title TEXT NOT NULL, publication_date DATE, url TEXT, description TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ
);
CREATE INDEX articles_profile_id_idx ON articles(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE skills (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 name TEXT NOT NULL, category TEXT, proficiency TEXT, years_experience NUMERIC(5,2),
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT skills_years_nonnegative CHECK (years_experience IS NULL OR years_experience >= 0)
);
CREATE INDEX skills_profile_id_idx ON skills(profile_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX skills_profile_name_unique_idx ON skills(profile_id, lower(name)) WHERE deleted_at IS NULL;

CREATE TABLE research_interests (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 name TEXT NOT NULL, description TEXT, priority INTEGER,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT research_interest_priority_nonnegative CHECK (priority IS NULL OR priority >= 0)
);
CREATE INDEX research_interests_profile_id_idx ON research_interests(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE career_goals (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 title TEXT NOT NULL, description TEXT NOT NULL, goal_type TEXT, priority INTEGER,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT career_goal_priority_nonnegative CHECK (priority IS NULL OR priority >= 0)
);
CREATE INDEX career_goals_profile_id_idx ON career_goals(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE certifications (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 name TEXT NOT NULL, issuer TEXT, issue_date DATE, expiration_date DATE, credential_id TEXT, url TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT certification_dates_valid CHECK (expiration_date IS NULL OR issue_date IS NULL OR expiration_date >= issue_date)
);
CREATE INDEX certifications_profile_id_idx ON certifications(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE awards (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 title TEXT NOT NULL, issuer TEXT, award_date DATE, description TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ
);
CREATE INDEX awards_profile_id_idx ON awards(profile_id) WHERE deleted_at IS NULL;

CREATE TABLE volunteer_experiences (
 id UUID PRIMARY KEY, profile_id UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
 organization TEXT NOT NULL, role TEXT NOT NULL, start_date DATE, end_date DATE, is_current BOOLEAN NOT NULL DEFAULT FALSE, description TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
 CONSTRAINT volunteering_dates_valid CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX volunteer_experiences_profile_id_idx ON volunteer_experiences(profile_id) WHERE deleted_at IS NULL;
