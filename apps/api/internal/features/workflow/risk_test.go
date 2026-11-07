package workflow

import (
	"context"
	"errors"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
	"math"
	"testing"
	"time"
)

func riskFloat(v float64) *float64 { return &v }
func riskString(v string) *string  { return &v }
func TestRiskMeanAndUnknowns(t *testing.T) {
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	factors := []RiskVariableRequest{{VariableName: "eligibility", Value: riskFloat(20), Rationale: "Evidence", ObservedAt: now}, {VariableName: "supervisor_accessibility", Value: riskFloat(80), Rationale: "Evidence", ObservedAt: now}, {VariableName: "funding", Rationale: "Unknown", ObservedAt: now}}
	score, err := validateRiskVariables(factors, now)
	if err != nil || score == nil || *score != 50 {
		t.Fatalf("score=%v err=%v", score, err)
	}
	factors = factors[2:]
	score, err = validateRiskVariables(factors, now)
	if err != nil || score != nil {
		t.Fatalf("all unknown must remain unscored")
	}
}
func TestRiskValidationAndTemporalPredicates(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)
	base := RiskVariableRequest{VariableName: "supervisor_accessibility", Value: riskFloat(25), Rationale: "Faculty response", ObservedAt: now}
	cases := map[string]func(*RiskVariableRequest){
		"negative": func(v *RiskVariableRequest) { v.Value = riskFloat(-1) }, "over100": func(v *RiskVariableRequest) { v.Value = riskFloat(101) }, "nan": func(v *RiskVariableRequest) { v.Value = riskFloat(math.NaN()) }, "infinite": func(v *RiskVariableRequest) { v.Value = riskFloat(math.Inf(1)) },
		"future evidence": func(v *RiskVariableRequest) { v.ObservedAt = future }, "zero evidence": func(v *RiskVariableRequest) { v.ObservedAt = time.Time{} }, "invalid name": func(v *RiskVariableRequest) { v.VariableName = "Supervisor Risk" }, "no rationale": func(v *RiskVariableRequest) { v.Rationale = " " },
		"no time predicate": func(v *RiskVariableRequest) { v.TimeVariable = true; v.ValidUntil = &future }, "no expiry": func(v *RiskVariableRequest) {
			v.TimeVariable = true
			v.TimePredicate = riskString("While positions remain available")
		}, "invalid expiry": func(v *RiskVariableRequest) { v.ValidUntil = &past }, "unsafe url": func(v *RiskVariableRequest) { v.SourceURL = riskString("javascript:alert(1)") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			v := base
			change(&v)
			if _, err := validateRiskVariables([]RiskVariableRequest{v}, now); err == nil {
				t.Fatal("invalid factor accepted")
			}
		})
	}
	if _, err := validateRiskVariables([]RiskVariableRequest{base, base}, now); err == nil {
		t.Fatal("duplicate variable counted twice")
	}
	base.TimeVariable = true
	base.TimePredicate = riskString("While this faculty member accepts MSc students")
	base.ValidUntil = &future
	if _, err := validateRiskVariables([]RiskVariableRequest{base}, now); err != nil {
		t.Fatal(err)
	}
}
func TestRiskHistoryStalenessDoesNotRecomputeHistoricalScore(t *testing.T) {
	expiry := time.Now().UTC()
	score := 45.0
	v := ProposalRiskAssessment{RiskFactor: &score, Variables: []ProposalRiskVariable{{VariableName: "supervisor", Value: riskFloat(45), ValidUntil: &expiry, Blocking: true}, {VariableName: "funding"}}}
	before := riskResponse(v, expiry.Add(-time.Second))
	after := riskResponse(v, expiry)
	if before.Stale || !after.Stale || after.UnknownCount != 1 || after.BlockingCount != 1 || *after.RiskFactor != 45 {
		t.Fatalf("unexpected freshness or historical score: %+v", after)
	}
}

type riskMemory struct {
	Repository
	proposal *ApplicationProposal
	saved    []*ProposalRiskAssessment
}

func (r *riskMemory) GetProposal(context.Context, uuid.UUID) (*ApplicationProposal, error) {
	return r.proposal, nil
}
func (r *riskMemory) ListProposalSources(context.Context, uuid.UUID) ([]ProposalSourceRecord, error) {
	return nil, nil
}
func (r *riskMemory) CreateRiskAssessment(_ context.Context, v *ProposalRiskAssessment) error {
	r.saved = append(r.saved, v)
	return nil
}
func TestRiskOwnershipScopeAndAppendOnlySnapshots(t *testing.T) {
	owner, other, id, pid := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	repo := &riskMemory{proposal: &ApplicationProposal{Base: Base{ID: id}, UserID: owner}}
	profiles := &fakeProfiles{item: &profile.ApplicantProfile{Base: profile.Base{ID: pid}, UserID: owner}}
	service := &Service{repo: repo, profiles: profiles, now: func() time.Time { return now }}
	q := CreateRiskAssessmentRequest{ProfileID: pid, Notes: "Test assumptions", Variables: []RiskVariableRequest{{VariableName: "eligibility", Value: riskFloat(20), Rationale: "Supported", ObservedAt: now}}}
	foreign := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &other})
	if _, err := service.CreateProposalRisk(foreign, id, q); !errors.Is(err, ErrApplicationProposalNotFound) {
		t.Fatalf("cross-owner assessment accepted: %v", err)
	}
	agent := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeResearch: {}}})
	if _, err := service.CreateProposalRisk(agent, id, q); !errors.Is(err, principal.ErrAgentScopeRequired) {
		t.Fatalf("agent without profiles scope accepted: %v", err)
	}
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &owner})
	first, err := service.CreateProposalRisk(ctx, id, q)
	if err != nil {
		t.Fatal(err)
	}
	q.Variables[0].Value = riskFloat(80)
	second, err := service.CreateProposalRisk(ctx, id, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.saved) != 2 || first.ID == second.ID || first.ProfileSnapshotID == second.ProfileSnapshotID || *first.RiskFactor != 20 || *second.RiskFactor != 80 {
		t.Fatal("history/snapshot overwritten")
	}
	profiles.item.UserID = other
	if _, err := service.CreateProposalRisk(ctx, id, q); !errors.Is(err, profile.ErrProfileNotFound) {
		t.Fatalf("foreign profile accepted: %v", err)
	}
}
