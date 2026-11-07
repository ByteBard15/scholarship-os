package workflow

import (
	"context"
	"encoding/json"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/google/uuid"
)

var riskNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,99}$`)

func validateRiskVariables(items []RiskVariableRequest, now time.Time) (*float64, error) {
	if len(items) == 0 || len(items) > 30 {
		return nil, validationError("provide 1 to 30 risk variables")
	}
	seen := map[string]bool{}
	total := 0.0
	count := 0
	for _, v := range items {
		if !riskNamePattern.MatchString(v.VariableName) || seen[v.VariableName] {
			return nil, validationError("variableName must be unique lower_snake_case")
		}
		seen[v.VariableName] = true
		if strings.TrimSpace(v.Rationale) == "" || len(v.Rationale) > 4000 {
			return nil, validationError("each variable requires a rationale within 4000 characters")
		}
		if v.Value != nil {
			if math.IsNaN(*v.Value) || math.IsInf(*v.Value, 0) || *v.Value < 0 || *v.Value > 100 {
				return nil, validationError("risk values must be finite numbers between 0 and 100")
			}
			total += *v.Value
			count++
		}
		if v.ObservedAt.IsZero() || v.ObservedAt.After(now) {
			return nil, validationError("observedAt is required and cannot be in the future")
		}
		if v.TimeVariable && (v.ValidUntil == nil || v.TimePredicate == nil || strings.TrimSpace(*v.TimePredicate) == "") {
			return nil, validationError("time-variable factors require validUntil and a timePredicate")
		}
		if v.ValidUntil != nil && !v.ValidUntil.After(v.ObservedAt) {
			return nil, validationError("validUntil must be after observedAt")
		}
		if v.TimePredicate != nil && len(*v.TimePredicate) > 2000 {
			return nil, validationError("timePredicate exceeds 2000 characters")
		}
		if v.SourceURL != nil {
			u, e := url.Parse(*v.SourceURL)
			if e != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") || len(*v.SourceURL) > 2000 {
				return nil, validationError("sourceUrl must be an HTTP(S) URL without credentials")
			}
		}
	}
	if count == 0 {
		return nil, nil
	}
	score := total / float64(count)
	return &score, nil
}
func (s *Service) riskAccess(ctx context.Context, id uuid.UUID, write bool) (*ApplicationProposal, error) {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if actor.Type == principal.PrincipalAgent {
		if !actor.HasScope(principal.ScopeResearch) {
			return nil, principal.ErrAgentScopeRequired
		}
		if write && !actor.HasScope(principal.ScopeProfiles) {
			return nil, principal.ErrAgentScopeRequired
		}
	}
	proposal, err := s.repo.GetProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, proposal.UserID) != nil {
		return nil, ErrApplicationProposalNotFound
	}
	return proposal, nil
}
func (s *Service) ListProposalRisk(ctx context.Context, id uuid.UUID) ([]RiskAssessmentResponse, error) {
	if _, err := s.riskAccess(ctx, id, false); err != nil {
		return nil, err
	}
	items, err := s.repo.ListRiskAssessments(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]RiskAssessmentResponse, 0, len(items))
	for _, v := range items {
		out = append(out, riskResponse(v, s.now()))
	}
	return out, nil
}
func (s *Service) CreateProposalRisk(ctx context.Context, id uuid.UUID, q CreateRiskAssessmentRequest) (*RiskAssessmentResponse, error) {
	proposal, err := s.riskAccess(ctx, id, true)
	if err != nil {
		return nil, err
	}
	now := s.now()
	score, err := validateRiskVariables(q.Variables, now)
	if err != nil {
		return nil, err
	}
	if q.ProfileID == uuid.Nil || strings.TrimSpace(q.Notes) == "" || len(q.Notes) > 10000 {
		return nil, validationError("profileId and notes are required")
	}
	p, err := s.profiles.Get(ctx, q.ProfileID)
	if err != nil {
		return nil, err
	}
	if p.UserID != proposal.UserID {
		return nil, profile.ErrProfileNotFound
	}
	by, reason := "risk_assessment", "Credentials at proposal risk computation"
	snapshot, err := s.profiles.CreateSnapshot(ctx, q.ProfileID, profile.CreateSnapshotRequest{CreatedBy: &by, Reason: &reason})
	if err != nil {
		return nil, err
	}
	actor, _ := principal.PrincipalFromContext(ctx)
	recordedBy := actor.UserID
	if actor.Type == principal.PrincipalAgent {
		recordedBy = actor.AgentID
	}
	sources, err := s.repo.ListProposalSources(ctx, id)
	if err != nil {
		return nil, err
	}
	contextJSON, err := json.Marshal(proposalResponse(proposal, sources))
	if err != nil {
		return nil, err
	}
	v := ProposalRiskAssessment{ID: uuid.New(), ProposalID: id, ProfileSnapshotID: snapshot.ID, ProposalSnapshot: profile.JSON(contextJSON), RiskFactor: score, Methodology: "mean_v1", Notes: strings.TrimSpace(q.Notes), RecordedBy: recordedBy, RecordedByType: string(actor.Type), ComputedAt: now, CreatedAt: now}
	for _, f := range q.Variables {
		v.Variables = append(v.Variables, ProposalRiskVariable{ID: uuid.New(), AssessmentID: v.ID, VariableName: f.VariableName, Value: f.Value, Rationale: strings.TrimSpace(f.Rationale), SourceURL: f.SourceURL, ObservedAt: f.ObservedAt.UTC(), TimeVariable: f.TimeVariable, TimePredicate: f.TimePredicate, ValidUntil: f.ValidUntil, Blocking: f.Blocking})
	}
	if err = s.repo.CreateRiskAssessment(ctx, &v); err != nil {
		return nil, err
	}
	appLogger.Info(ctx, "proposal risk assessment recorded", "proposal_id", id, "assessment_id", v.ID)
	out := riskResponse(v, now)
	return &out, nil
}
