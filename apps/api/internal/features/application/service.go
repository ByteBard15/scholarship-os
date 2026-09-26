package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/scholarship-os/apps/api/internal/features/catalog"
	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/example/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type ProfileReader interface {
	Get(context.Context, uuid.UUID) (*profile.ApplicantProfile, error)
	ResolveEffectiveProfile(context.Context, uuid.UUID) (*profile.EffectiveProfileResponse, error)
	ListEvidence(context.Context, uuid.UUID) ([]profile.ProfileEvidence, error)
}
type CatalogReader interface {
	GetInstitution(context.Context, uuid.UUID) (*catalog.Institution, error)
	GetProgramme(context.Context, uuid.UUID) (*catalog.Programme, error)
	GetScholarship(context.Context, uuid.UUID) (*catalog.Scholarship, error)
}
type PreparationReader interface {
	ApplicationPreparation(context.Context, uuid.UUID) (PreparationComponents, error)
}
type PreparationComponents struct {
	FieldsCompleted, FieldsTotal                                          int
	QuestionnairesCompleted, QuestionnairesTotal                          int
	PendingInformation, RequiredQuestionsAnswered, RequiredQuestionsTotal int
}

type Service struct {
	repo        Repository
	profiles    ProfileReader
	catalog     CatalogReader
	preparation PreparationReader
	now         func() time.Time
}

func NewService(r Repository, p ProfileReader, c CatalogReader) *Service {
	return &Service{repo: r, profiles: p, catalog: c, now: func() time.Time { return time.Now().UTC() }}
}
func (s *Service) SetPreparationReader(reader PreparationReader) { s.preparation = reader }
func (s *Service) audit(a *Application, action string, entityType string, entityID uuid.UUID) *profile.AuditLog {
	metadata, _ := json.Marshal(map[string]any{"application_id": a.ID})
	return &profile.AuditLog{UserID: &a.UserID, ProfileID: &a.ApplicantProfileID, Action: action, EntityType: &entityType, EntityID: &entityID, Metadata: profile.JSON(metadata)}
}

func (s *Service) Create(ctx context.Context, q CreateApplicationRequest) (*Application, error) {
	parent, e := s.profiles.Get(ctx, q.ParentProfileID)
	if e != nil {
		if errors.Is(e, profile.ErrProfileNotFound) {
			return nil, ErrApplicationProfileRequired
		}
		return nil, e
	}
	if parent.ProfileType != profile.ProfileTypeMaster && parent.ProfileType != profile.ProfileTypeDomain {
		return nil, ErrApplicationProfileRequired
	}
	if e = s.validateCatalog(ctx, q.InstitutionID, q.ProgrammeID, q.ScholarshipID); e != nil {
		return nil, e
	}
	p := &profile.ApplicantProfile{UserID: parent.UserID, Name: q.Name + " Application Profile", ProfileType: profile.ProfileTypeApplication, ParentProfileID: &parent.ID}
	a := &Application{UserID: parent.UserID, InstitutionID: q.InstitutionID, ProgrammeID: q.ProgrammeID, ScholarshipID: q.ScholarshipID, Name: strings.TrimSpace(q.Name), Intake: q.Intake, IntakeYear: q.IntakeYear, Country: q.Country, Status: StatusDiscovered, Priority: q.Priority, Notes: q.Notes, ResearchStatus: ResearchNotStarted}
	createTasks := q.CreateDefaultTasks == nil || *q.CreateDefaultTasks
	var tasks []ApplicationTask
	if createTasks {
		tasks = defaultTasks()
	}
	entityType := "application"
	audit := &profile.AuditLog{UserID: &parent.UserID, Action: "application.created", EntityType: &entityType, Metadata: profile.JSON(`{"application_profile_created":true}`)}
	if e = s.repo.CreateWithProfile(ctx, p, a, tasks, audit); e != nil {
		return nil, e
	}
	return s.repo.Get(ctx, a.ID)
}
func defaultTasks() []ApplicationTask {
	names := []struct{ title, kind string }{{"Research opportunity", "research"}, {"Verify eligibility", "research"}, {"Verify deadlines", "research"}, {"Review required documents", "document"}, {"Prepare application profile", "profile"}, {"Prepare CV", "cv"}, {"Request references", "reference"}, {"Review and submit application", "submission"}}
	v := make([]ApplicationTask, len(names))
	for i, n := range names {
		order := i + 1
		kind := n.kind
		source := "system"
		v[i] = ApplicationTask{Title: n.title, Status: "todo", SortOrder: &order, TaskType: &kind, Source: &source}
	}
	return v
}
func (s *Service) validateCatalog(ctx context.Context, i, p, sch *uuid.UUID) error {
	if i != nil {
		if _, e := s.catalog.GetInstitution(ctx, *i); e != nil {
			return e
		}
	}
	if p != nil {
		v, e := s.catalog.GetProgramme(ctx, *p)
		if e != nil {
			return e
		}
		if i != nil && v.InstitutionID != *i {
			return newValidationError("programme does not belong to the selected institution")
		}
	}
	if sch != nil {
		if _, e := s.catalog.GetScholarship(ctx, *sch); e != nil {
			return e
		}
	}
	return nil
}
func (s *Service) List(ctx context.Context, f Filters) ([]Application, error) {
	if actor, ok := principal.PrincipalFromContext(ctx); ok && !actor.IsSystem() {
		if actor.UserID == nil {
			return nil, principal.ErrForbidden
		}
		f.UserID = actor.UserID
	}
	return s.repo.List(ctx, f)
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Application, error) {
	value, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := principal.EnforceOwner(ctx, value.UserID); err != nil {
		return nil, ErrApplicationNotFound
	}
	return value, nil
}
func (s *Service) Update(ctx context.Context, id uuid.UUID, q UpdateApplicationRequest) (*Application, error) {
	a, e := s.repo.Get(ctx, id)
	if e != nil {
		return nil, e
	}
	if principal.EnforceOwner(ctx, a.UserID) != nil {
		return nil, ErrApplicationNotFound
	}
	old := a.Status
	if q.Name != nil {
		a.Name = strings.TrimSpace(*q.Name)
	}
	if q.InstitutionID != nil {
		a.InstitutionID = q.InstitutionID
	}
	if q.ProgrammeID != nil {
		a.ProgrammeID = q.ProgrammeID
	}
	if q.ScholarshipID != nil {
		a.ScholarshipID = q.ScholarshipID
	}
	if q.Intake != nil {
		a.Intake = q.Intake
	}
	if q.IntakeYear != nil {
		a.IntakeYear = q.IntakeYear
	}
	if q.Country != nil {
		a.Country = q.Country
	}
	if q.Priority != nil {
		a.Priority = q.Priority
	}
	if q.Notes != nil {
		a.Notes = q.Notes
	}
	if e = s.validateCatalog(ctx, a.InstitutionID, a.ProgrammeID, a.ScholarshipID); e != nil {
		return nil, e
	}
	var audit *profile.AuditLog
	if q.Status != nil && *q.Status != a.Status {
		if !q.Status.Valid() || (!q.ManualStatusOverride && !CanTransition(a.Status, *q.Status)) {
			return nil, ErrInvalidApplicationStatus
		}
		a.Status = *q.Status
		audit = s.audit(a, "application.status.changed", "application", a.ID)
		meta, _ := json.Marshal(map[string]any{"from": old, "to": a.Status, "manual_override": q.ManualStatusOverride})
		audit.Metadata = profile.JSON(meta)
	}
	if e = s.repo.Update(ctx, a, audit); e != nil {
		return nil, e
	}
	return s.repo.Get(ctx, id)
}
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	a, e := s.repo.Get(ctx, id)
	if e != nil {
		return e
	}
	return s.repo.Delete(ctx, a)
}

func DeadlineUrgency(now time.Time, deadline *time.Time, precision string) string {
	if deadline == nil || (precision != "exact" && precision != "day") {
		return "unknown"
	}
	n := now.UTC()
	d := deadline.UTC()
	ny, nm, nd := n.Date()
	dy, dm, dd := d.Date()
	today := time.Date(ny, nm, nd, 0, 0, 0, 0, time.UTC)
	target := time.Date(dy, dm, dd, 0, 0, 0, 0, time.UTC)
	days := int(target.Sub(today).Hours() / 24)
	switch {
	case days < 0:
		return "past_due"
	case days == 0:
		return "today"
	case days == 1:
		return "tomorrow"
	case days <= 3:
		return "within_3_days"
	case days <= 7:
		return "within_7_days"
	case days <= 14:
		return "within_14_days"
	case days <= 30:
		return "within_30_days"
	default:
		return "later"
	}
}
func eligibility(reqs []ApplicationRequirement) EligibilitySummary {
	v := EligibilitySummary{}
	for _, r := range reqs {
		if !r.IsMandatory {
			continue
		}
		v.MandatoryTotal++
		switch r.Status {
		case "satisfied", "waived", "not_applicable":
			v.Satisfied++
		case "not_satisfied":
			v.NotSatisfied++
		default:
			v.Unknown++
		}
	}
	return v
}
func readiness(a *Application, reqs []ApplicationRequirement, deadlines []ApplicationDeadline, tasks []ApplicationTask) ReadinessSummary {
	done, total := 0, 5
	if a.ResearchStatus == ResearchComplete {
		done++
	}
	for _, d := range deadlines {
		if d.DeadlineAt != nil && (d.DatePrecision == "exact" || d.DatePrecision == "day") {
			done++
			break
		}
	}
	allReviewed := len(reqs) > 0
	for _, r := range reqs {
		if r.IsMandatory && (r.Status == "unknown" || r.Status == "not_started") {
			allReviewed = false
			break
		}
	}
	if allReviewed {
		done++
	}
	if a.ApplicantProfileID != uuid.Nil {
		done++
	}
	if len(tasks) > 0 {
		all := true
		for _, t := range tasks {
			if t.Status != "done" && t.Status != "cancelled" {
				all = false
				break
			}
		}
		if all {
			done++
		}
	}
	state := "in_progress"
	if done == total {
		state = "ready"
	} else if done == 0 {
		state = "not_started"
	}
	return ReadinessSummary{state, done, total, "Operational readiness only; this is not an admission or scholarship likelihood score."}
}
func (s *Service) Detail(ctx context.Context, id uuid.UUID) (*ApplicationDetailResponse, error) {
	d, e := s.repo.GetDetail(ctx, id)
	if e != nil {
		return nil, e
	}
	if principal.EnforceOwner(ctx, d.Application.UserID) != nil {
		return nil, ErrApplicationNotFound
	}
	return detailResponse(d, s.now()), nil
}
func detailResponse(d *ApplicationDetail, now time.Time) *ApplicationDetailResponse {
	v := &ApplicationDetailResponse{Application: toApplication(d.Application), Requirements: make([]RequirementResponse, len(d.Requirements)), Deadlines: make([]DeadlineResponse, len(d.Deadlines)), Funding: make([]FundingResponse, len(d.Funding)), Contacts: make([]ContactResponse, len(d.Contacts)), Supervisors: make([]SupervisorResponse, len(d.Supervisors)), URLs: make([]URLResponse, len(d.URLs)), Tasks: make([]TaskResponse, len(d.Tasks))}
	for i := range d.Requirements {
		v.Requirements[i] = toRequirement(&d.Requirements[i])
	}
	for i := range d.Deadlines {
		v.Deadlines[i] = toDeadline(&d.Deadlines[i], now)
	}
	for i := range d.Funding {
		v.Funding[i] = toFunding(&d.Funding[i])
	}
	for i := range d.Contacts {
		v.Contacts[i] = toContact(&d.Contacts[i])
	}
	for i := range d.Supervisors {
		v.Supervisors[i] = toSupervisor(&d.Supervisors[i])
	}
	for i := range d.URLs {
		v.URLs[i] = toURL(&d.URLs[i])
	}
	for i := range d.Tasks {
		v.Tasks[i] = toTask(&d.Tasks[i])
	}
	if d.LatestResearchRun != nil {
		x := toResearchRun(d.LatestResearchRun)
		v.LatestResearchRun = &x
	}
	return v
}
func (s *Service) Summary(ctx context.Context, id uuid.UUID) (*ApplicationSummaryResponse, error) {
	d, e := s.repo.GetDetail(ctx, id)
	if e != nil {
		return nil, e
	}
	if principal.EnforceOwner(ctx, d.Application.UserID) != nil {
		return nil, ErrApplicationNotFound
	}
	v := &ApplicationSummaryResponse{ApplicationResponse: toApplication(d.Application), Readiness: readiness(d.Application, d.Requirements, d.Deadlines, d.Tasks)}
	if s.preparation != nil {
		components, prepErr := s.preparation.ApplicationPreparation(ctx, id)
		if prepErr != nil {
			return nil, prepErr
		}
		v.Readiness = mergePreparationReadiness(v.Readiness, components)
	}
	for _, x := range d.Deadlines {
		if x.DeadlineAt != nil && (x.DatePrecision == "exact" || x.DatePrecision == "day") {
			y := toDeadline(&x, s.now())
			v.NextDeadline = &y
			break
		}
	}
	v.TaskCompletion.Total = len(d.Tasks)
	for _, t := range d.Tasks {
		if t.Status == "done" || t.Status == "cancelled" {
			v.TaskCompletion.Completed++
		}
	}
	for _, r := range d.Requirements {
		if r.IsMandatory {
			v.RequirementCompletion.Total++
			if r.Status == "satisfied" || r.Status == "waived" || r.Status == "not_applicable" {
				v.RequirementCompletion.Completed++
			}
		}
	}
	return v, nil
}
func (s *Service) Readiness(ctx context.Context, id uuid.UUID) (ReadinessSummary, error) {
	d, e := s.repo.GetDetail(ctx, id)
	if e != nil {
		return ReadinessSummary{}, e
	}
	if principal.EnforceOwner(ctx, d.Application.UserID) != nil {
		return ReadinessSummary{}, ErrApplicationNotFound
	}
	result := readiness(d.Application, d.Requirements, d.Deadlines, d.Tasks)
	if s.preparation != nil {
		components, prepErr := s.preparation.ApplicationPreparation(ctx, id)
		if prepErr != nil {
			return ReadinessSummary{}, prepErr
		}
		result = mergePreparationReadiness(result, components)
	}
	return result, nil
}

func mergePreparationReadiness(result ReadinessSummary, components PreparationComponents) ReadinessSummary {
	result.TotalItems += 4
	if components.FieldsTotal > 0 && components.FieldsCompleted == components.FieldsTotal {
		result.CompletedItems++
	}
	if components.QuestionnairesTotal > 0 && components.QuestionnairesCompleted == components.QuestionnairesTotal {
		result.CompletedItems++
	}
	if components.RequiredQuestionsTotal > 0 && components.RequiredQuestionsAnswered == components.RequiredQuestionsTotal {
		result.CompletedItems++
	}
	if components.PendingInformation == 0 {
		result.CompletedItems++
	}
	if result.CompletedItems == result.TotalItems {
		result.Status = "ready"
	} else if result.CompletedItems == 0 {
		result.Status = "not_started"
	} else {
		result.Status = "in_progress"
	}
	return result
}
func (s *Service) Eligibility(ctx context.Context, id uuid.UUID) (EligibilitySummary, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return EligibilitySummary{}, err
	}
	v, e := s.repo.ListRequirements(ctx, id)
	if e != nil {
		return EligibilitySummary{}, e
	}
	return eligibility(v), nil
}

func (s *Service) ListRequirements(c context.Context, a uuid.UUID) ([]ApplicationRequirement, error) {
	return s.repo.ListRequirements(c, a)
}
func (s *Service) CreateRequirement(c context.Context, a uuid.UUID, q RequirementRequest) (*ApplicationRequirement, error) {
	if _, e := s.repo.Get(c, a); e != nil {
		return nil, e
	}
	mandatory := true
	if q.IsMandatory != nil {
		mandatory = *q.IsMandatory
	}
	status := "unknown"
	if q.Status != nil {
		status = *q.Status
	}
	evidence := false
	if q.EvidenceRequired != nil {
		evidence = *q.EvidenceRequired
	}
	v := &ApplicationRequirement{ApplicationID: a, Category: q.Category, Title: q.Title, Description: q.Description, IsMandatory: mandatory, Status: status, DueDate: q.DueDate, SourceURL: q.SourceURL, EvidenceRequired: evidence, Notes: q.Notes, SortOrder: q.SortOrder}
	app, _ := s.repo.Get(c, a)
	return v, s.repo.CreateRequirement(c, v, s.audit(app, "requirement.created", "application_requirement", uuid.Nil))
}
func (s *Service) GetRequirement(c context.Context, a, id uuid.UUID) (*ApplicationRequirement, error) {
	return s.repo.GetRequirement(c, a, id)
}
func (s *Service) UpdateRequirement(c context.Context, a, id uuid.UUID, q RequirementRequest) (*ApplicationRequirement, error) {
	v, e := s.repo.GetRequirement(c, a, id)
	if e != nil {
		return nil, e
	}
	v.Category = q.Category
	v.Title = q.Title
	v.Description = q.Description
	if q.IsMandatory != nil {
		v.IsMandatory = *q.IsMandatory
	}
	if q.Status != nil {
		v.Status = *q.Status
	}
	v.DueDate = q.DueDate
	v.SourceURL = q.SourceURL
	if q.EvidenceRequired != nil {
		v.EvidenceRequired = *q.EvidenceRequired
	}
	v.Notes = q.Notes
	v.SortOrder = q.SortOrder
	app, _ := s.repo.Get(c, a)
	return v, s.repo.UpdateRequirement(c, v, s.audit(app, "requirement.updated", "application_requirement", v.ID))
}
func (s *Service) DeleteRequirement(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetRequirement(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteRequirement(c, v)
}

func (s *Service) ListDeadlines(c context.Context, a uuid.UUID) ([]ApplicationDeadline, error) {
	return s.repo.ListDeadlines(c, a)
}
func (s *Service) CreateDeadline(c context.Context, a uuid.UUID, q DeadlineRequest) (*ApplicationDeadline, error) {
	app, e := s.repo.Get(c, a)
	if e != nil {
		return nil, e
	}
	if q.DeadlineAt == nil && (q.DatePrecision == "exact" || q.DatePrecision == "day") {
		return nil, newValidationError("exact or day precision requires deadline_at")
	}
	hard := false
	if q.IsHardDeadline != nil {
		hard = *q.IsHardDeadline
	}
	v := &ApplicationDeadline{ApplicationID: a, DeadlineType: q.DeadlineType, Title: q.Title, DeadlineAt: q.DeadlineAt, Timezone: q.Timezone, DatePrecision: q.DatePrecision, RawDeadlineText: q.RawDeadlineText, IsHardDeadline: hard, SourceURL: q.SourceURL, VerifiedAt: q.VerifiedAt, Notes: q.Notes}
	e = s.repo.CreateDeadline(c, v, s.audit(app, "deadline.created", "application_deadline", uuid.Nil))
	return v, e
}
func (s *Service) UpdateDeadline(c context.Context, a, id uuid.UUID, q DeadlineRequest) (*ApplicationDeadline, error) {
	v, e := s.repo.GetDeadline(c, a, id)
	if e != nil {
		return nil, e
	}
	if q.DeadlineAt == nil && (q.DatePrecision == "exact" || q.DatePrecision == "day") {
		return nil, newValidationError("exact or day precision requires deadline_at")
	}
	v.DeadlineType = q.DeadlineType
	v.Title = q.Title
	v.DeadlineAt = q.DeadlineAt
	v.Timezone = q.Timezone
	v.DatePrecision = q.DatePrecision
	v.RawDeadlineText = q.RawDeadlineText
	if q.IsHardDeadline != nil {
		v.IsHardDeadline = *q.IsHardDeadline
	}
	v.SourceURL = q.SourceURL
	v.VerifiedAt = q.VerifiedAt
	v.Notes = q.Notes
	app, _ := s.repo.Get(c, a)
	return v, s.repo.UpdateDeadline(c, v, s.audit(app, "deadline.updated", "application_deadline", v.ID))
}
func (s *Service) DeleteDeadline(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetDeadline(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteDeadline(c, v)
}
func (s *Service) UpcomingDeadlines(c context.Context, days int, appID *uuid.UUID, status *Status) ([]ApplicationDeadline, error) {
	if days < 1 || days > 365 {
		return nil, newValidationError("days must be between 1 and 365")
	}
	n := s.now()
	var userID *uuid.UUID
	if actor, ok := principal.PrincipalFromContext(c); ok && !actor.IsSystem() {
		userID = actor.UserID
	}
	return s.repo.UpcomingDeadlines(c, n, n.AddDate(0, 0, days), appID, userID, status)
}

func (s *Service) ListFunding(c context.Context, a uuid.UUID) ([]ApplicationFunding, error) {
	return s.repo.ListFunding(c, a)
}
func fundingFrom(a uuid.UUID, q FundingRequest) *ApplicationFunding {
	return &ApplicationFunding{ApplicationID: a, FundingType: q.FundingType, Currency: q.Currency, Amount: q.Amount, AmountPeriod: q.AmountPeriod, TuitionCoverage: q.TuitionCoverage, StipendAmount: q.StipendAmount, StipendPeriod: q.StipendPeriod, TravelCoverage: q.TravelCoverage, InsuranceCoverage: q.InsuranceCoverage, AccommodationCoverage: q.AccommodationCoverage, OtherBenefits: q.OtherBenefits, Conditions: q.Conditions, SourceURL: q.SourceURL}
}
func (s *Service) CreateFunding(c context.Context, a uuid.UUID, q FundingRequest) (*ApplicationFunding, error) {
	if _, e := s.repo.Get(c, a); e != nil {
		return nil, e
	}
	v := fundingFrom(a, q)
	return v, s.repo.CreateFunding(c, v)
}
func (s *Service) UpdateFunding(c context.Context, a, id uuid.UUID, q FundingRequest) (*ApplicationFunding, error) {
	v, e := s.repo.GetFunding(c, a, id)
	if e != nil {
		return nil, e
	}
	n := fundingFrom(a, q)
	n.Base = v.Base
	n.SoftDelete = v.SoftDelete
	n.SourceID = v.SourceID
	return n, s.repo.UpdateFunding(c, n)
}
func (s *Service) DeleteFunding(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetFunding(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteFunding(c, v)
}

func (s *Service) ListContacts(c context.Context, a uuid.UUID) ([]ApplicationContact, error) {
	return s.repo.ListContacts(c, a)
}
func contactFrom(a uuid.UUID, q ContactRequest) *ApplicationContact {
	return &ApplicationContact{ApplicationID: a, Name: q.Name, Role: q.Role, Email: q.Email, Phone: q.Phone, Organization: q.Organization, ContactType: q.ContactType, URL: q.URL, Notes: q.Notes}
}
func (s *Service) CreateContact(c context.Context, a uuid.UUID, q ContactRequest) (*ApplicationContact, error) {
	if _, e := s.repo.Get(c, a); e != nil {
		return nil, e
	}
	v := contactFrom(a, q)
	return v, s.repo.CreateContact(c, v)
}
func (s *Service) UpdateContact(c context.Context, a, id uuid.UUID, q ContactRequest) (*ApplicationContact, error) {
	v, e := s.repo.GetContact(c, a, id)
	if e != nil {
		return nil, e
	}
	n := contactFrom(a, q)
	n.Base = v.Base
	n.SoftDelete = v.SoftDelete
	n.SourceID = v.SourceID
	return n, s.repo.UpdateContact(c, n)
}
func (s *Service) DeleteContact(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetContact(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteContact(c, v)
}

func (s *Service) ListSupervisors(c context.Context, a uuid.UUID) ([]ApplicationSupervisor, error) {
	return s.repo.ListSupervisors(c, a)
}
func supervisorFrom(a uuid.UUID, q SupervisorRequest) *ApplicationSupervisor {
	return &ApplicationSupervisor{ApplicationID: a, Name: q.Name, Title: q.Title, Department: q.Department, Institution: q.Institution, Email: q.Email, ProfileURL: q.ProfileURL, ResearchAreas: q.ResearchAreas, ContactStatus: q.ContactStatus, Notes: q.Notes}
}
func (s *Service) CreateSupervisor(c context.Context, a uuid.UUID, q SupervisorRequest) (*ApplicationSupervisor, error) {
	if _, e := s.repo.Get(c, a); e != nil {
		return nil, e
	}
	v := supervisorFrom(a, q)
	return v, s.repo.CreateSupervisor(c, v)
}
func (s *Service) UpdateSupervisor(c context.Context, a, id uuid.UUID, q SupervisorRequest) (*ApplicationSupervisor, error) {
	v, e := s.repo.GetSupervisor(c, a, id)
	if e != nil {
		return nil, e
	}
	n := supervisorFrom(a, q)
	n.Base = v.Base
	n.SoftDelete = v.SoftDelete
	n.SourceID = v.SourceID
	return n, s.repo.UpdateSupervisor(c, n)
}
func (s *Service) DeleteSupervisor(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetSupervisor(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteSupervisor(c, v)
}

func (s *Service) ListURLs(c context.Context, a uuid.UUID) ([]ApplicationURL, error) {
	return s.repo.ListURLs(c, a)
}
func (s *Service) CreateURL(c context.Context, a uuid.UUID, q URLRequest) (*ApplicationURL, error) {
	if _, e := s.repo.Get(c, a); e != nil {
		return nil, e
	}
	v := &ApplicationURL{ApplicationID: a, URLType: q.URLType, Label: q.Label, URL: q.URL, IsOfficial: q.IsOfficial}
	return v, s.repo.CreateURL(c, v)
}
func (s *Service) DeleteURL(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetURL(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteURL(c, v)
}

func (s *Service) ListTasks(c context.Context, a uuid.UUID) ([]ApplicationTask, error) {
	return s.repo.ListTasks(c, a)
}
func (s *Service) validateTaskParent(c context.Context, a, current uuid.UUID, parent *uuid.UUID) error {
	if parent == nil {
		return nil
	}
	if *parent == current {
		return ErrTaskParentCycle
	}
	seen := map[uuid.UUID]bool{current: true}
	id := parent
	for depth := 0; id != nil; depth++ {
		if depth > 10 || seen[*id] {
			return ErrTaskParentCycle
		}
		seen[*id] = true
		p, e := s.repo.GetTask(c, a, *id)
		if e != nil {
			return e
		}
		id = p.ParentTaskID
	}
	return nil
}
func (s *Service) CreateTask(c context.Context, a uuid.UUID, q TaskRequest) (*ApplicationTask, error) {
	if _, e := s.repo.Get(c, a); e != nil {
		return nil, e
	}
	id := uuid.New()
	if e := s.validateTaskParent(c, a, id, q.ParentTaskID); e != nil {
		return nil, e
	}
	status := "todo"
	if q.Status != nil {
		status = *q.Status
	}
	v := &ApplicationTask{ApplicationID: a, ParentTaskID: q.ParentTaskID, Title: q.Title, Description: q.Description, Status: status, Priority: q.Priority, DueAt: q.DueAt, SortOrder: q.SortOrder, TaskType: q.TaskType, Source: q.Source}
	v.ID = id
	if status == "done" {
		n := s.now()
		v.CompletedAt = &n
	}
	return v, s.repo.CreateTask(c, v)
}
func (s *Service) UpdateTask(c context.Context, a, id uuid.UUID, q TaskRequest) (*ApplicationTask, error) {
	v, e := s.repo.GetTask(c, a, id)
	if e != nil {
		return nil, e
	}
	if e = s.validateTaskParent(c, a, id, q.ParentTaskID); e != nil {
		return nil, e
	}
	v.ParentTaskID = q.ParentTaskID
	v.Title = q.Title
	v.Description = q.Description
	if q.Status != nil {
		v.Status = *q.Status
		if v.Status == "done" && v.CompletedAt == nil {
			n := s.now()
			v.CompletedAt = &n
		} else if v.Status != "done" {
			v.CompletedAt = nil
		}
	}
	v.Priority = q.Priority
	v.DueAt = q.DueAt
	v.SortOrder = q.SortOrder
	v.TaskType = q.TaskType
	v.Source = q.Source
	return v, s.repo.UpdateTask(c, v, nil)
}
func (s *Service) DeleteTask(c context.Context, a, id uuid.UUID) error {
	v, e := s.repo.GetTask(c, a, id)
	if e != nil {
		return e
	}
	return s.repo.DeleteTask(c, v)
}
func (s *Service) SetTaskCompletion(c context.Context, a, id uuid.UUID, complete bool) (*ApplicationTask, error) {
	v, e := s.repo.GetTask(c, a, id)
	if e != nil {
		return nil, e
	}
	action := "application.task.reopened"
	if complete {
		v.Status = "done"
		n := s.now()
		v.CompletedAt = &n
		action = "application.task.completed"
	} else {
		v.Status = "todo"
		v.CompletedAt = nil
	}
	app, e := s.repo.Get(c, a)
	if e != nil {
		return nil, e
	}
	return v, s.repo.UpdateTask(c, v, s.audit(app, action, "application_task", v.ID))
}

func (s *Service) EvidenceSuggestions(c context.Context, a, requirementID uuid.UUID) ([]EvidenceSuggestion, error) {
	app, e := s.repo.Get(c, a)
	if e != nil {
		return nil, e
	}
	req, e := s.repo.GetRequirement(c, a, requirementID)
	if e != nil {
		return nil, e
	}
	effective, e := s.profiles.ResolveEffectiveProfile(c, app.ApplicantProfileID)
	if e != nil {
		return nil, e
	}
	var source []profile.SectionResponse
	entityType := req.Category
	reason := "Potential profile evidence; user review is required."
	switch req.Category {
	case "academic", "admission", "transcript":
		source = effective.Education
		entityType = "education"
	case "work_experience":
		source = effective.Employment
		entityType = "employment"
	case "cv", "document":
		source = append(source, effective.Education...)
		source = append(source, effective.Employment...)
		source = append(source, effective.Projects...)
	case "research_proposal":
		source = effective.ResearchInterests
		entityType = "research_interests"
	default:
		source = append(source, effective.Certifications...)
		source = append(source, effective.Awards...)
	}
	out := make([]EvidenceSuggestion, 0, len(source))
	for _, v := range source {
		label := v.Title
		if label == nil {
			label = v.Name
		}
		if label == nil {
			label = v.Institution
		}
		if label == nil {
			label = v.Organization
		}
		text := "Profile entry"
		if label != nil {
			text = *label
		}
		out = append(out, EvidenceSuggestion{EntityType: entityType, EntityID: v.ID, Label: text, Reason: reason})
	}
	return out, nil
}
func (s *Service) ListEvidence(c context.Context, a, requirementID uuid.UUID) ([]RequirementEvidence, error) {
	if _, e := s.repo.GetRequirement(c, a, requirementID); e != nil {
		return nil, e
	}
	return s.repo.ListEvidence(c, requirementID)
}
func (s *Service) CreateEvidence(c context.Context, a, requirementID uuid.UUID, q EvidenceRequest) (*RequirementEvidence, error) {
	app, e := s.repo.Get(c, a)
	if e != nil {
		return nil, e
	}
	if _, e = s.repo.GetRequirement(c, a, requirementID); e != nil {
		return nil, e
	}
	effective, e := s.profiles.ResolveEffectiveProfile(c, app.ApplicantProfileID)
	if e != nil {
		return nil, e
	}
	sections := map[string][]profile.SectionResponse{"education": effective.Education, "employment": effective.Employment, "projects": effective.Projects, "publications": effective.Publications, "articles": effective.Articles, "skills": effective.Skills, "research_interests": effective.ResearchInterests, "career_goals": effective.CareerGoals, "certifications": effective.Certifications, "awards": effective.Awards, "volunteering": effective.Volunteering}
	entityFound := false
	if q.EntityType == "personal_info" && effective.PersonalInfo != nil && effective.PersonalInfo.ID == q.EntityID {
		entityFound = true
	}
	for _, item := range sections[q.EntityType] {
		if item.ID == q.EntityID {
			entityFound = true
			break
		}
	}
	if !entityFound {
		return nil, newValidationError("evidence entity is not present in the effective application profile")
	}
	if q.EvidenceID != nil {
		matched := false
		for _, lineage := range effective.Lineage {
			available, err := s.profiles.ListEvidence(c, lineage.ID)
			if err != nil {
				return nil, err
			}
			for _, item := range available {
				if item.ID == *q.EvidenceID && item.EntityID == q.EntityID {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return nil, newValidationError("profile evidence does not support the selected effective-profile entity")
		}
	}
	v := &RequirementEvidence{RequirementID: requirementID, ProfileID: app.ApplicantProfileID, EntityType: q.EntityType, EntityID: q.EntityID, EvidenceID: q.EvidenceID, Status: q.Status, Notes: q.Notes}
	audit := s.audit(app, "requirement.evidence.linked", "requirement_evidence", uuid.Nil)
	return v, s.repo.CreateEvidence(c, v, audit)
}

func (s *Service) Dashboard(c context.Context) (*DashboardResponse, error) {
	var userID *uuid.UUID
	if actor, ok := principal.PrincipalFromContext(c); ok && !actor.IsSystem() {
		userID = actor.UserID
	}
	d, e := s.repo.Dashboard(c, userID)
	if e != nil {
		return nil, e
	}
	v := &DashboardResponse{ActiveApplications: make([]ApplicationResponse, len(d.Applications)), UpcomingDeadlines: make([]DeadlineResponse, len(d.Deadlines)), TasksDueSoon: make([]TaskResponse, len(d.Tasks)), ResearchPendingReview: d.PendingResearch, UnresolvedResearchConflicts: d.ResearchConflicts, IncompleteMandatoryRequirements: d.IncompleteRequirements, StatusSummary: map[Status]int{}}
	for i := range d.Applications {
		v.ActiveApplications[i] = toApplication(&d.Applications[i])
		v.StatusSummary[d.Applications[i].Status]++
		if d.Applications[i].ResearchStatus == ResearchNotStarted || d.Applications[i].ResearchStatus == ResearchStale || d.Applications[i].ResearchStatus == ResearchFailed {
			v.ApplicationsNeedingResearch++
		}
	}
	for i := range d.Deadlines {
		v.UpcomingDeadlines[i] = toDeadline(&d.Deadlines[i], s.now())
	}
	for i := range d.Tasks {
		v.TasksDueSoon[i] = toTask(&d.Tasks[i])
	}
	return v, nil
}

func requireApplicationID(v uuid.UUID) error {
	if v == uuid.Nil {
		return fmt.Errorf("application ID is required")
	}
	return nil
}
func isValidationError(e error) bool { var v validationError; return errors.As(e, &v) }
