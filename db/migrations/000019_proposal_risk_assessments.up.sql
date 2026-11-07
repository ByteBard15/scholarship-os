CREATE TABLE proposal_risk_assessments (
 id UUID PRIMARY KEY,
 proposal_id UUID NOT NULL REFERENCES application_proposals(id) ON DELETE CASCADE,
 profile_snapshot_id UUID NOT NULL REFERENCES profile_snapshots(id) ON DELETE RESTRICT,
 proposal_snapshot JSONB NOT NULL,
 risk_factor DOUBLE PRECISION CHECK (risk_factor BETWEEN 0 AND 100),
 methodology TEXT NOT NULL DEFAULT 'mean_v1',
 notes TEXT NOT NULL,
 recorded_by UUID,
 recorded_by_type TEXT NOT NULL,
 computed_at TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX proposal_risk_assessments_proposal_idx ON proposal_risk_assessments(proposal_id, computed_at DESC, id);
CREATE TABLE proposal_risk_variables (
 id UUID PRIMARY KEY,
 assessment_id UUID NOT NULL REFERENCES proposal_risk_assessments(id) ON DELETE CASCADE,
 variable_name TEXT NOT NULL,
 value DOUBLE PRECISION CHECK (value BETWEEN 0 AND 100),
 rationale TEXT NOT NULL,
 source_url TEXT,
 observed_at TIMESTAMPTZ NOT NULL,
 time_variable BOOLEAN NOT NULL DEFAULT FALSE,
 time_predicate TEXT,
 valid_until TIMESTAMPTZ,
 blocking BOOLEAN NOT NULL DEFAULT FALSE,
 UNIQUE(assessment_id, variable_name),
 CHECK (valid_until IS NULL OR valid_until > observed_at),
 CHECK (NOT time_variable OR (valid_until IS NOT NULL AND time_predicate IS NOT NULL AND length(trim(time_predicate)) > 0))
);
CREATE INDEX proposal_risk_variables_assessment_idx ON proposal_risk_variables(assessment_id);
