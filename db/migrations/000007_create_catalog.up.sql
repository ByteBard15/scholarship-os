CREATE TABLE institutions (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    short_name TEXT,
    institution_type TEXT,
    country TEXT NOT NULL,
    city TEXT,
    website_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_institutions_country ON institutions(country) WHERE deleted_at IS NULL;
CREATE INDEX idx_institutions_name ON institutions(name) WHERE deleted_at IS NULL;

CREATE TABLE programmes (
    id UUID PRIMARY KEY,
    institution_id UUID NOT NULL REFERENCES institutions(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    degree_level TEXT,
    field_of_study TEXT,
    faculty TEXT,
    department TEXT,
    duration_months INTEGER CHECK (duration_months IS NULL OR duration_months > 0),
    mode TEXT,
    language TEXT,
    programme_url TEXT,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_programmes_institution_id ON programmes(institution_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_programmes_degree_level ON programmes(degree_level) WHERE deleted_at IS NULL;

CREATE TABLE scholarships (
    id UUID PRIMARY KEY,
    institution_id UUID REFERENCES institutions(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    provider_name TEXT,
    description TEXT,
    country TEXT,
    degree_level TEXT,
    scholarship_type TEXT,
    official_url TEXT,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_scholarships_institution_id ON scholarships(institution_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_scholarships_country ON scholarships(country) WHERE deleted_at IS NULL;
CREATE INDEX idx_scholarships_degree_level ON scholarships(degree_level) WHERE deleted_at IS NULL;
