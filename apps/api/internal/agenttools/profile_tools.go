package agenttools

import (
	"context"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

// ProfileAgentTools exposes owner-authorized profile maintenance without
// Application bootstrap or generic persistence access.
type ProfileAgentTools interface {
	UpdatePersonalInfo(context.Context, uuid.UUID, profile.PersonalInfoRequest) error
	AddProfileSection(context.Context, profile.SectionKind, uuid.UUID, profile.SectionRequest) error
	UpdateProfileSection(context.Context, profile.SectionKind, uuid.UUID, uuid.UUID, profile.SectionRequest) error
	ListProfiles(context.Context) ([]ProfileTailoringView, error)
	ResolveProfile(context.Context, uuid.UUID) (*profile.EffectiveProfileResponse, error)
	CreateDomainProfile(context.Context, profile.AgentCreateDomainRequest) (*ProfileTailoringView, error)
	TailorProfile(context.Context, uuid.UUID, profile.AgentUpdateProfileRequest) (*ProfileTailoringView, error)
	TailorDescription(context.Context, uuid.UUID, profile.CreateOverrideRequest) error
}

type ProfileTailoringView struct {
	ID              uuid.UUID
	ParentProfileID *uuid.UUID
	ProfileType     profile.ProfileType
	Name            string
	Headline        *string
	Summary         *string
}

type ControlledProfileTools struct{ profiles *profile.Service }

func NewProfileTools(profiles *profile.Service) *ControlledProfileTools {
	return &ControlledProfileTools{profiles: profiles}
}
func tailoringView(p *profile.ApplicantProfile) *ProfileTailoringView {
	return &ProfileTailoringView{ID: p.ID, ParentProfileID: p.ParentProfileID, ProfileType: p.ProfileType, Name: p.Name, Headline: p.Headline, Summary: p.Summary}
}
func (t *ControlledProfileTools) ListProfiles(ctx context.Context) ([]ProfileTailoringView, error) {
	rows, err := t.profiles.AgentListProfiles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ProfileTailoringView, 0, len(rows))
	for i := range rows {
		out = append(out, *tailoringView(&rows[i]))
	}
	return out, nil
}
func (t *ControlledProfileTools) ResolveProfile(ctx context.Context, id uuid.UUID) (*profile.EffectiveProfileResponse, error) {
	return t.profiles.AgentEffectiveProfile(ctx, id)
}
func (t *ControlledProfileTools) CreateDomainProfile(ctx context.Context, req profile.AgentCreateDomainRequest) (*ProfileTailoringView, error) {
	p, err := t.profiles.AgentCreateDomain(ctx, req)
	if err != nil {
		return nil, err
	}
	return tailoringView(p), nil
}
func (t *ControlledProfileTools) TailorProfile(ctx context.Context, id uuid.UUID, req profile.AgentUpdateProfileRequest) (*ProfileTailoringView, error) {
	p, err := t.profiles.AgentUpdateProfile(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return tailoringView(p), nil
}
func (t *ControlledProfileTools) TailorDescription(ctx context.Context, id uuid.UUID, req profile.CreateOverrideRequest) error {
	_, err := t.profiles.AgentCreateDescriptionOverride(ctx, id, req)
	return err
}

var _ ProfileAgentTools = (*ControlledProfileTools)(nil)

func requireProfileWriter(ctx context.Context) error {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return err
	}
	if actor.IsSystem() {
		return nil
	}
	if actor.Type != principal.PrincipalAgent {
		return principal.ErrAgentRequired
	}
	if !actor.HasScope(principal.ScopeProfiles) {
		return principal.ErrAgentScopeRequired
	}
	return nil
}
func (t *ControlledProfileTools) UpdatePersonalInfo(ctx context.Context, id uuid.UUID, req profile.PersonalInfoRequest) error {
	if err := requireProfileWriter(ctx); err != nil {
		return err
	}
	_, err := t.profiles.PatchPersonalInfo(ctx, id, req)
	return err
}
func (t *ControlledProfileTools) AddProfileSection(ctx context.Context, kind profile.SectionKind, id uuid.UUID, req profile.SectionRequest) error {
	if err := requireProfileWriter(ctx); err != nil {
		return err
	}
	_, err := t.profiles.CreateSection(ctx, kind, id, req)
	return err
}
func (t *ControlledProfileTools) UpdateProfileSection(ctx context.Context, kind profile.SectionKind, id, sectionID uuid.UUID, req profile.SectionRequest) error {
	if err := requireProfileWriter(ctx); err != nil {
		return err
	}
	_, err := t.profiles.UpdateSection(ctx, kind, id, sectionID, req)
	return err
}
