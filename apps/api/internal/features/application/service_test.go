package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
)

type fakeApplicationRepo struct {
	Repository
	app            *Application
	createdProfile *profile.ApplicantProfile
	tasks          []ApplicationTask
	requirements   []ApplicationRequirement
	updatedAudit   *profile.AuditLog
}

func (f *fakeApplicationRepo) CreateWithProfile(_ context.Context, p *profile.ApplicantProfile, a *Application, tasks []ApplicationTask, _ *profile.AuditLog) error {
	p.ID = uuid.New()
	a.ID = uuid.New()
	a.ApplicantProfileID = p.ID
	f.createdProfile = p
	f.app = a
	f.tasks = tasks
	return nil
}
func (f *fakeApplicationRepo) Get(_ context.Context, id uuid.UUID) (*Application, error) {
	if f.app == nil || f.app.ID != id {
		return nil, ErrApplicationNotFound
	}
	return f.app, nil
}
func (f *fakeApplicationRepo) Update(_ context.Context, a *Application, audit *profile.AuditLog) error {
	f.app = a
	f.updatedAudit = audit
	return nil
}
func (f *fakeApplicationRepo) GetTask(_ context.Context, appID, id uuid.UUID) (*ApplicationTask, error) {
	for i := range f.tasks {
		if f.tasks[i].ApplicationID == appID && f.tasks[i].ID == id {
			return &f.tasks[i], nil
		}
	}
	return nil, ErrTaskNotFound
}
func (f *fakeApplicationRepo) UpdateTask(_ context.Context, v *ApplicationTask, a *profile.AuditLog) error {
	for i := range f.tasks {
		if f.tasks[i].ID == v.ID {
			f.tasks[i] = *v
		}
	}
	f.updatedAudit = a
	return nil
}
func (f *fakeApplicationRepo) GetRequirement(_ context.Context, appID, id uuid.UUID) (*ApplicationRequirement, error) {
	for i := range f.requirements {
		if f.requirements[i].ApplicationID == appID && f.requirements[i].ID == id {
			return &f.requirements[i], nil
		}
	}
	return nil, ErrRequirementNotFound
}

type fakeProfiles struct {
	parent    *profile.ApplicantProfile
	effective *profile.EffectiveProfileResponse
}

func (f fakeProfiles) Get(_ context.Context, id uuid.UUID) (*profile.ApplicantProfile, error) {
	if f.parent == nil || f.parent.ID != id {
		return nil, profile.ErrProfileNotFound
	}
	return f.parent, nil
}
func (f fakeProfiles) ResolveEffectiveProfile(_ context.Context, _ uuid.UUID) (*profile.EffectiveProfileResponse, error) {
	return f.effective, nil
}
func (f fakeProfiles) ListEvidence(context.Context, uuid.UUID) ([]profile.ProfileEvidence, error) {
	return nil, nil
}
func TestCreateApplicationCreatesOwnedProfileAndTasks(t *testing.T) {
	parent := &profile.ApplicantProfile{UserID: uuid.New(), Name: "Biomedical", ProfileType: profile.ProfileTypeDomain}
	parent.ID = uuid.New()
	repo := &fakeApplicationRepo{}
	svc := NewService(repo, fakeProfiles{parent: parent}, nil)
	created, e := svc.Create(context.Background(), CreateApplicationRequest{Name: "Example 2027", ParentProfileID: parent.ID})
	if e != nil {
		t.Fatal(e)
	}
	if repo.createdProfile == nil || repo.createdProfile.ProfileType != profile.ProfileTypeApplication {
		t.Fatal("application profile was not created")
	}
	if repo.createdProfile.UserID != parent.UserID || created.UserID != parent.UserID {
		t.Fatal("application ownership was not inherited from parent profile")
	}
	if repo.createdProfile.ParentProfileID == nil || *repo.createdProfile.ParentProfileID != parent.ID {
		t.Fatal("application profile parent was not retained")
	}
	if len(repo.tasks) != 8 {
		t.Fatalf("got %d default tasks, want 8", len(repo.tasks))
	}
}
func TestCreateApplicationRejectsApplicationParent(t *testing.T) {
	parent := &profile.ApplicantProfile{UserID: uuid.New(), ProfileType: profile.ProfileTypeApplication}
	parent.ID = uuid.New()
	svc := NewService(&fakeApplicationRepo{}, fakeProfiles{parent: parent}, nil)
	_, e := svc.Create(context.Background(), CreateApplicationRequest{Name: "Nested", ParentProfileID: parent.ID})
	if !errors.Is(e, ErrApplicationProfileRequired) {
		t.Fatalf("got %v", e)
	}
}
func TestStatusTransitions(t *testing.T) {
	if !CanTransition(StatusDiscovered, StatusResearching) {
		t.Fatal("expected discovered -> researching")
	}
	if CanTransition(StatusDiscovered, StatusAccepted) {
		t.Fatal("unexpected discovered -> accepted")
	}
}
func TestDeadlineUrgencyAndPrecision(t *testing.T) {
	now := time.Date(2027, 1, 1, 10, 0, 0, 0, time.UTC)
	tomorrow := now.AddDate(0, 0, 1)
	if got := DeadlineUrgency(now, &tomorrow, "exact"); got != "tomorrow" {
		t.Fatalf("got %s", got)
	}
	if got := DeadlineUrgency(now, &tomorrow, "month"); got != "unknown" {
		t.Fatalf("vague deadline got %s", got)
	}
}
func TestEligibilitySummary(t *testing.T) {
	v := eligibility([]ApplicationRequirement{{IsMandatory: true, Status: "satisfied"}, {IsMandatory: true, Status: "not_satisfied"}, {IsMandatory: true, Status: "unknown"}, {IsMandatory: false, Status: "unknown"}})
	if v.MandatoryTotal != 3 || v.Satisfied != 1 || v.NotSatisfied != 1 || v.Unknown != 1 {
		t.Fatalf("unexpected summary: %+v", v)
	}
}

func TestTaskCompletionReopenAndParentCycle(t *testing.T) {
	app := &Application{UserID: uuid.New(), ApplicantProfileID: uuid.New()}
	app.ID = uuid.New()
	parentID, childID := uuid.New(), uuid.New()
	repo := &fakeApplicationRepo{app: app, tasks: []ApplicationTask{{Base: Base{ID: parentID}, ApplicationID: app.ID, ParentTaskID: &childID, Title: "Parent", Status: "todo"}, {Base: Base{ID: childID}, ApplicationID: app.ID, Title: "Child", Status: "todo"}}}
	svc := NewService(repo, fakeProfiles{}, nil)
	v, e := svc.SetTaskCompletion(context.Background(), app.ID, childID, true)
	if e != nil || v.Status != "done" || v.CompletedAt == nil {
		t.Fatalf("complete failed: %+v %v", v, e)
	}
	v, e = svc.SetTaskCompletion(context.Background(), app.ID, childID, false)
	if e != nil || v.Status != "todo" || v.CompletedAt != nil {
		t.Fatalf("reopen failed: %+v %v", v, e)
	}
	if e = svc.validateTaskParent(context.Background(), app.ID, childID, &parentID); !errors.Is(e, ErrTaskParentCycle) {
		t.Fatalf("got %v", e)
	}
}

func TestRequirementEvidenceSuggestionsUseEffectiveProfile(t *testing.T) {
	app := &Application{ApplicantProfileID: uuid.New()}
	app.ID = uuid.New()
	req := ApplicationRequirement{ApplicationID: app.ID, Category: "academic", Title: "Degree"}
	req.ID = uuid.New()
	educationID := uuid.New()
	effective := &profile.EffectiveProfileResponse{FullProfileResponse: profile.FullProfileResponse{Education: []profile.SectionResponse{{ID: educationID, Institution: ptrString("Example University")}}}}
	svc := NewService(&fakeApplicationRepo{app: app, requirements: []ApplicationRequirement{req}}, fakeProfiles{effective: effective}, nil)
	v, e := svc.EvidenceSuggestions(context.Background(), app.ID, req.ID)
	if e != nil {
		t.Fatal(e)
	}
	if len(v) != 1 || v[0].EntityType != "education" || v[0].EntityID != educationID {
		t.Fatalf("unexpected suggestions: %+v", v)
	}
}

type fakeResearchRepo struct {
	ResearchRepository
	runs     []ResearchRun
	sources  []ResearchSource
	findings []ResearchFinding
	applied  AppliedResearch
}

func (f *fakeResearchRepo) CreateRun(_ context.Context, v *ResearchRun, _ *Application) error {
	v.ID = uuid.New()
	f.runs = append(f.runs, *v)
	return nil
}
func (f *fakeResearchRepo) PersistResult(_ context.Context, r *ResearchRun, _ *Application, s []ResearchSource, findings []ResearchFinding) error {
	f.runs[0] = *r
	f.sources = s
	f.findings = findings
	return nil
}
func (f *fakeResearchRepo) FailRun(context.Context, *ResearchRun, *Application, error) error {
	return nil
}
func (f *fakeResearchRepo) GetRun(_ context.Context, a, id uuid.UUID) (*ResearchRun, error) {
	for i := range f.runs {
		if f.runs[i].ID == id && f.runs[i].ApplicationID != nil && *f.runs[i].ApplicationID == a {
			return &f.runs[i], nil
		}
	}
	return nil, ErrResearchRunNotFound
}
func (f *fakeResearchRepo) ListFindings(_ context.Context, run uuid.UUID) ([]ResearchFinding, error) {
	return f.findings, nil
}
func (f *fakeResearchRepo) GetFinding(_ context.Context, run, id uuid.UUID) (*ResearchFinding, error) {
	for i := range f.findings {
		if f.findings[i].ID == id {
			return &f.findings[i], nil
		}
	}
	return nil, ErrResearchFindingNotFound
}
func (f *fakeResearchRepo) UpdateFinding(_ context.Context, v *ResearchFinding) error {
	for i := range f.findings {
		if f.findings[i].ID == v.ID {
			f.findings[i] = *v
		}
	}
	return nil
}
func (f *fakeResearchRepo) ApplyFindings(_ context.Context, _ *ResearchRun, _ *Application, b AppliedResearch, _ *profile.AuditLog) error {
	f.applied = b
	return nil
}

func TestConflictingFindingsAreRetained(t *testing.T) {
	run := &ResearchRun{}
	run.ID = uuid.New()
	app := &Application{}
	app.ID = uuid.New()
	a, b := time.Now(), time.Now().AddDate(0, 0, 1)
	result := &ResearchResult{Deadlines: []DeadlineCandidate{{Data: DeadlineRequest{DeadlineType: "scholarship", Title: "Close", DeadlineAt: &a, DatePrecision: "exact"}}, {Data: DeadlineRequest{DeadlineType: "scholarship", Title: "Close", DeadlineAt: &b, DatePrecision: "exact"}}}}
	_, findings, e := buildResearchRecords(run, app, result)
	if e != nil {
		t.Fatal(e)
	}
	if len(findings) != 2 || findings[0].VerificationStatus != "conflicting" || findings[1].VerificationStatus != "conflicting" {
		t.Fatalf("conflicts not retained: %+v", findings)
	}
}
