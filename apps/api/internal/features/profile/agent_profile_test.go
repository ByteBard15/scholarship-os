package profile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type agentWorkflow struct {
	WorkflowRepository
	overrides []ProfileOverride
}

func (m *agentWorkflow) ListOverrides(_ context.Context, id uuid.UUID) ([]ProfileOverride, error) {
	out := []ProfileOverride{}
	for _, o := range m.overrides {
		if o.ProfileID == id {
			out = append(out, o)
		}
	}
	return out, nil
}
func (m *agentWorkflow) CreateOverride(_ context.Context, o *ProfileOverride) error {
	o.ID = uuid.New()
	m.overrides = append(m.overrides, *o)
	return nil
}
func (m *agentWorkflow) CreateAudit(context.Context, *AuditLog) error { return nil }
func profileAgentContext(owner uuid.UUID, scope bool) context.Context {
	scopes := map[string]struct{}{}
	if scope {
		scopes[principal.ScopeProfiles] = struct{}{}
	}
	return principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: scopes})
}
func TestAgentProfileMaintenancePreservesOwnership(t *testing.T) {
	owner, other := uuid.New(), uuid.New()
	repo := newMemoryProfiles()
	wf := &agentWorkflow{}
	s := NewService(repo, nil, memoryUsers{item: &user.User{ID: owner}}, WithWorkflowRepository(wf))
	original := "Canonical summary"
	master := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: owner, Name: "Master", ProfileType: ProfileTypeMaster, Summary: &original, IsDefault: true}
	repo.items[master.ID] = master
	ctx := profileAgentContext(owner, true)
	tailored := "Biomedical signal processing interests"
	domain, err := s.AgentCreateDomain(ctx, AgentCreateDomainRequest{ParentProfileID: master.ID, Name: "Biomedical", Summary: &tailored})
	if err != nil || domain.ProfileType != ProfileTypeDomain || domain.IsDefault || domain.ParentProfileID == nil || *domain.ParentProfileID != master.ID {
		t.Fatalf("bad derived profile: %+v %v", domain, err)
	}
	next := "Neural signals and computation"
	if _, err := s.AgentUpdateProfile(ctx, domain.ID, AgentUpdateProfileRequest{Summary: &next}); err != nil {
		t.Fatal(err)
	}
	effective, err := s.AgentEffectiveProfile(ctx, domain.ID)
	if err != nil || effective.Summary == nil || *effective.Summary != next || len(effective.Lineage) != 2 {
		t.Fatalf("effective profile: %+v %v", effective, err)
	}
	if *repo.items[master.ID].Summary != original {
		t.Fatal("derived update changed master")
	}
	if _, err := s.AgentUpdateProfile(ctx, master.ID, AgentUpdateProfileRequest{Summary: &next}); err != nil {
		t.Fatal("authorized Master edit rejected", err)
	}
	if *repo.items[master.ID].Summary != next {
		t.Fatal("Master edit not saved")
	}
	for _, id := range []uuid.UUID{master.ID, domain.ID} {
		if _, err := s.AgentEffectiveProfile(profileAgentContext(other, true), id); err != ErrProfileNotFound {
			t.Fatalf("foreign read allowed: %v", err)
		}
		if _, err := s.AgentUpdateProfile(profileAgentContext(other, true), id, AgentUpdateProfileRequest{Summary: &next}); err != ErrProfileNotFound {
			t.Fatalf("foreign write allowed: %v", err)
		}
	}
	if _, err := s.AgentCreateDomain(ctx, AgentCreateDomainRequest{ParentProfileID: domain.ID, Name: "Invalid nesting"}); err == nil {
		t.Fatal("domain nested below non-master")
	}
	if _, err := s.AgentCreateDomain(profileAgentContext(other, true), AgentCreateDomainRequest{ParentProfileID: master.ID, Name: "Foreign"}); err != ErrProfileNotFound {
		t.Fatalf("foreign parent allowed: %v", err)
	}
	if _, err := s.AgentListProfiles(profileAgentContext(owner, false)); err != principal.ErrAgentScopeRequired {
		t.Fatalf("missing scope accepted: %v", err)
	}
	if _, err := s.AgentUpdateProfile(context.Background(), domain.ID, AgentUpdateProfileRequest{Summary: &next}); err != principal.ErrAuthRequired {
		t.Fatalf("unauthenticated write: %v", err)
	}
	application := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: owner, Name: "Owned application", ProfileType: ProfileTypeApplication, ParentProfileID: &domain.ID}
	repo.items[application.ID] = application
	if _, err := s.AgentUpdateProfile(ctx, application.ID, AgentUpdateProfileRequest{Summary: &next}); err != nil {
		t.Fatal(err)
	}
	rows, err := s.AgentListProfiles(ctx)
	if err != nil || len(rows) != 3 {
		t.Fatalf("owned listing: %d %v", len(rows), err)
	}
}
func TestAgentDescriptionOverridesCannotChangeFacts(t *testing.T) {
	owner := uuid.New()
	repo := newMemoryProfiles()
	wf := &agentWorkflow{}
	s := NewService(repo, nil, memoryUsers{item: &user.User{ID: owner}}, WithWorkflowRepository(wf))
	ctx := profileAgentContext(owner, true)
	master := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: owner, Name: "Master", ProfileType: ProfileTypeMaster}
	repo.items[master.ID] = master
	domain := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: owner, Name: "Biomedical", ProfileType: ProfileTypeDomain, ParentProfileID: &master.ID}
	repo.items[domain.ID] = domain
	entity := uuid.New()
	reason := "Emphasise supported DSP coursework"
	value, _ := json.Marshal("Supported wording")
	req := CreateOverrideRequest{EntityType: "education", EntityID: &entity, FieldName: "grade", OverrideType: OverrideReplace, Value: value, Reason: &reason}
	if _, err := s.AgentCreateDescriptionOverride(ctx, domain.ID, req); err == nil {
		t.Fatal("grade modification allowed")
	}
	req.FieldName = "description"
	req.OverrideType = OverrideHide
	if _, err := s.AgentCreateDescriptionOverride(ctx, domain.ID, req); err == nil {
		t.Fatal("hiding facts allowed")
	}
	req.OverrideType = OverrideReplace
	if _, err := s.AgentCreateDescriptionOverride(ctx, master.ID, req); err == nil {
		t.Fatal("Master override allowed")
	}
	if _, err := s.AgentCreateDescriptionOverride(profileAgentContext(uuid.New(), true), domain.ID, req); err != ErrProfileNotFound {
		t.Fatal("foreign override allowed", err)
	}
	// The existing override resolver must reject a nonexistent source entity.
	if _, err := s.AgentCreateDescriptionOverride(ctx, domain.ID, req); err != ErrSectionNotFound {
		t.Fatalf("unresolved entity accepted: %v", err)
	}
	if len(wf.overrides) != 0 {
		t.Fatal("invalid overrides persisted")
	}
	original := "Original education wording"
	master.Education = []EducationHistory{{Base: Base{ID: entity}, ProfileID: master.ID, Institution: "University", Description: &original}}
	override, err := s.AgentCreateDescriptionOverride(ctx, domain.ID, req)
	if err != nil || override.Source == nil || *override.Source != "agent" {
		t.Fatalf("supported description rejected: %+v %v", override, err)
	}
	effective, err := s.AgentEffectiveProfile(ctx, domain.ID)
	if err != nil || len(effective.Education) != 1 || effective.Education[0].Description == nil || *effective.Education[0].Description != "Supported wording" {
		t.Fatalf("description not inherited and tailored: %+v %v", effective, err)
	}
	if *master.Education[0].Description != original {
		t.Fatal("canonical education wording changed")
	}
}
func TestAgentProfileHandlerRejectsPrivilegeFields(t *testing.T) {
	h := NewHandler(nil, validator.New())
	r := chi.NewRouter()
	r.Post("/agent/profiles", h.AgentCreateDomain)
	r.Patch("/agent/profiles/{profileID}", h.AgentUpdateProfile)
	for _, test := range []struct{ method, path, body string }{
		{"POST", "/agent/profiles", `{"name":"Malicious","parentProfileId":"` + uuid.NewString() + `","profileType":"master"}`},
		{"PATCH", "/agent/profiles/" + uuid.NewString(), `{"isDefault":true}`},
		{"PATCH", "/agent/profiles/" + uuid.NewString(), `{"userId":"` + uuid.NewString() + `"}`},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(test.method, test.path, strings.NewReader(test.body)))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("privilege field accepted: %d %s", w.Code, w.Body.String())
		}
	}
}

type agentEducationSections struct {
	SectionRepository
	row *EducationHistory
}

func (m *agentEducationSections) CreateEducation(_ context.Context, row *EducationHistory) error {
	row.ID = uuid.New()
	copy := *row
	m.row = &copy
	return nil
}
func (m *agentEducationSections) GetEducation(_ context.Context, pid, id uuid.UUID) (*EducationHistory, error) {
	if m.row == nil || m.row.ProfileID != pid || m.row.ID != id {
		return nil, ErrSectionNotFound
	}
	copy := *m.row
	return &copy, nil
}
func (m *agentEducationSections) UpdateEducation(_ context.Context, row *EducationHistory) error {
	copy := *row
	m.row = &copy
	return nil
}
func TestAgentMasterSectionMaintenance(t *testing.T) {
	owner := uuid.New()
	repo := newMemoryProfiles()
	master := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: owner, ProfileType: ProfileTypeMaster}
	other := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: owner, ProfileType: ProfileTypeMaster}
	repo.items[master.ID] = master
	repo.items[other.ID] = other
	sections := &agentEducationSections{}
	s := NewService(repo, sections, memoryUsers{item: &user.User{ID: owner}}, WithWorkflowRepository(&agentWorkflow{}))
	ctx := profileAgentContext(owner, true)
	institution, degree, field, description := "University", "BEng", "Electronic and Computer Engineering", "Verified project description"
	req := SectionRequest{Institution: &institution, Degree: &degree, FieldOfStudy: &field, Description: &description}
	if _, err := s.CreateSection(ctx, EducationKind, master.ID, req); err != nil {
		t.Fatal(err)
	}
	grade := "4.24"
	if _, err := s.UpdateSection(ctx, EducationKind, master.ID, sections.row.ID, SectionRequest{Grade: &grade}); err != nil {
		t.Fatal(err)
	}
	if sections.row.Grade == nil || *sections.row.Grade != grade {
		t.Fatal("Master education update not persisted")
	}
	if _, err := s.UpdateSection(ctx, EducationKind, other.ID, sections.row.ID, req); err != ErrSectionNotFound {
		t.Fatal("cross-profile section accepted", err)
	}
	for _, test := range []struct {
		ctx  context.Context
		want error
	}{
		{profileAgentContext(uuid.New(), true), ErrProfileNotFound},
		{profileAgentContext(owner, false), principal.ErrAgentScopeRequired},
	} {
		if _, err := s.CreateSection(test.ctx, EducationKind, master.ID, req); err != test.want {
			t.Fatalf("create guard: %v", err)
		}
		if _, err := s.UpdateSection(test.ctx, EducationKind, master.ID, sections.row.ID, req); err != test.want {
			t.Fatalf("update guard: %v", err)
		}
		if _, err := s.PatchPersonalInfo(test.ctx, master.ID, PersonalInfoRequest{}); err != test.want {
			t.Fatalf("personal guard: %v", err)
		}
	}
}
