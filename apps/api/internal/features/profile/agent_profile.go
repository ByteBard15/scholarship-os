package profile

import (
	"context"
	"encoding/json"
	"strings"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/google/uuid"
)

// Agent profile operations support owner-authorized Master maintenance.
// Application-profile creation belongs to approved-proposal bootstrap.
type AgentCreateDomainRequest struct {
	ParentProfileID uuid.UUID `json:"parentProfileId" validate:"required"`
	Name            string    `json:"name" validate:"required,max=250"`
	Headline        *string   `json:"headline" validate:"omitempty,max=500"`
	Summary         *string   `json:"summary" validate:"omitempty,max=20000"`
}
type AgentUpdateProfileRequest struct {
	Name     *string `json:"name" validate:"omitempty,min=1,max=250"`
	Headline *string `json:"headline" validate:"omitempty,max=500"`
	Summary  *string `json:"summary" validate:"omitempty,max=20000"`
}

func requireProfileAgent(ctx context.Context) (principal.Principal, error) {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return actor, err
	}
	if actor.IsSystem() {
		return actor, nil
	}
	if actor.Type != principal.PrincipalAgent {
		return actor, principal.ErrAgentRequired
	}
	if !actor.HasScope(principal.ScopeProfiles) {
		return actor, principal.ErrAgentScopeRequired
	}
	if actor.UserID == nil {
		return actor, principal.ErrForbidden
	}
	return actor, nil
}

func (s *Service) AgentListProfiles(ctx context.Context) ([]ApplicantProfile, error) {
	actor, err := requireProfileAgent(ctx)
	if err != nil {
		return nil, err
	}
	if actor.UserID == nil {
		return nil, newValidationError("a workspace-associated principal is required to list profiles")
	}
	return s.List(ctx, *actor.UserID)
}
func (s *Service) AgentEffectiveProfile(ctx context.Context, id uuid.UUID) (*EffectiveProfileResponse, error) {
	if _, err := requireProfileAgent(ctx); err != nil {
		return nil, err
	}
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.ResolveEffectiveProfile(ctx, id)
}
func (s *Service) AgentCreateDomain(ctx context.Context, req AgentCreateDomainRequest) (*ApplicantProfile, error) {
	if _, err := requireProfileAgent(ctx); err != nil {
		return nil, err
	}
	parent, err := s.Get(ctx, req.ParentProfileID)
	if err != nil {
		return nil, err
	}
	if parent.ProfileType != ProfileTypeMaster {
		return nil, newValidationError("domain profiles must inherit from an owned Master profile")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, newValidationError("name is required")
	}
	result, err := s.Create(ctx, parent.UserID, CreateProfileRequest{Name: req.Name, ProfileType: ProfileTypeDomain, ParentProfileID: &parent.ID, Headline: req.Headline, Summary: req.Summary})
	if err == nil {
		appLogger.Info(ctx, "agent created scholarship domain profile", "profile_id", result.ID, "parent_profile_id", parent.ID)
	}
	return result, err
}
func (s *Service) agentDerivedProfile(ctx context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	if _, err := requireProfileAgent(ctx); err != nil {
		return nil, err
	}
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.ProfileType == ProfileTypeMaster {
		return nil, newValidationError("agents may tailor derived profiles only; Master facts require human review")
	}
	return p, nil
}
func (s *Service) AgentUpdateProfile(ctx context.Context, id uuid.UUID, req AgentUpdateProfileRequest) (*ApplicantProfile, error) {
	if _, err := requireProfileAgent(ctx); err != nil {
		return nil, err
	}
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	if req.Name == nil && req.Headline == nil && req.Summary == nil {
		return nil, newValidationError("provide name, headline or summary")
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return nil, newValidationError("name cannot be blank")
	}
	result, err := s.Update(ctx, id, UpdateProfileRequest{Name: req.Name, Headline: req.Headline, Summary: req.Summary})
	if err == nil {
		s.recordAgentProfileWrite(ctx, id, "agent.profile.updated")
	}
	return result, err
}
func (s *Service) AgentCreateDescriptionOverride(ctx context.Context, id uuid.UUID, req CreateOverrideRequest) (*ProfileOverride, error) {
	if _, err := s.agentDerivedProfile(ctx, id); err != nil {
		return nil, err
	}
	// Only wording can be changed: IDs, grades, dates, authorship and other
	// factual fields are excluded from this endpoint's allowlist.
	kind := normalizeEntityType(req.EntityType)
	switch kind {
	case "education", "employment", "projects", "publications", "articles", "certifications", "awards", "volunteering", "research_interests", "career_goals":
	default:
		return nil, newValidationError("only descriptive section overrides are supported")
	}
	if req.EntityID == nil || req.FieldName != "description" || req.OverrideType != OverrideReplace {
		return nil, newValidationError("provide an entityId and a replace override for description")
	}
	var text string
	if json.Unmarshal(req.Value, &text) != nil || strings.TrimSpace(text) == "" {
		return nil, newValidationError("description must be a non-empty string")
	}
	if req.Reason == nil || strings.TrimSpace(*req.Reason) == "" {
		return nil, newValidationError("a reason identifying supporting profile facts is required")
	}
	source := "agent"
	req.Source = &source
	return s.CreateOverride(ctx, id, req)
}

// Existing user-facing services also enforce scoped ownership when invoked by
// an agent, so internal callers cannot bypass the dedicated route middleware.
func (s *Service) authorizeAgentProfileAccess(ctx context.Context, id uuid.UUID) error {
	actor, ok := principal.PrincipalFromContext(ctx)
	if !ok || actor.Type != principal.PrincipalAgent {
		return nil
	}
	if _, err := requireProfileAgent(ctx); err != nil {
		return err
	}
	_, err := s.Get(ctx, id)
	return err
}

func (s *Service) recordAgentProfileWrite(ctx context.Context, id uuid.UUID, action string) {
	actor, ok := principal.PrincipalFromContext(ctx)
	if !ok || actor.Type != principal.PrincipalAgent {
		return
	}
	appLogger.Info(ctx, "agent maintained profile", "profile_id", id, "action", action)
	if s.workflow != nil {
		if err := s.workflow.CreateAudit(ctx, &AuditLog{UserID: actor.UserID, ProfileID: &id, Action: action}); err != nil {
			appLogger.Error(ctx, "could not audit agent profile maintenance", "profile_id", id, "error", err)
		}
	}
}
