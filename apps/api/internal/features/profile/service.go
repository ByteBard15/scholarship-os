package profile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/example/scholarship-os/apps/api/internal/features/user"
	principal "github.com/example/scholarship-os/apps/api/pkg/auth"
	"github.com/example/scholarship-os/apps/api/pkg/filestore"
	"github.com/google/uuid"
)

type SectionKind string

const (
	EducationKind         SectionKind = "education"
	EmploymentKind        SectionKind = "employment"
	ProjectsKind          SectionKind = "projects"
	PublicationsKind      SectionKind = "publications"
	ArticlesKind          SectionKind = "articles"
	SkillsKind            SectionKind = "skills"
	ResearchInterestsKind SectionKind = "research-interests"
	CareerGoalsKind       SectionKind = "career-goals"
	CertificationsKind    SectionKind = "certifications"
	AwardsKind            SectionKind = "awards"
	VolunteeringKind      SectionKind = "volunteering"
)

type Service struct {
	profiles       Repository
	sections       SectionRepository
	users          user.Repository
	workflow       WorkflowRepository
	files          filestore.FileStore
	textExtractor  TextExtractor
	extractor      ProfileExtractor
	maxUploadBytes int64
}

type ServiceOption func(*Service)

func WithWorkflowRepository(repository WorkflowRepository) ServiceOption {
	return func(service *Service) { service.workflow = repository }
}
func WithImportWorkflow(files filestore.FileStore, textExtractor TextExtractor, extractor ProfileExtractor, maxUploadBytes int64) ServiceOption {
	return func(service *Service) {
		service.files = files
		service.textExtractor = textExtractor
		service.extractor = extractor
		service.maxUploadBytes = maxUploadBytes
	}
}

func NewService(profiles Repository, sections SectionRepository, users user.Repository, options ...ServiceOption) *Service {
	service := &Service{profiles: profiles, sections: sections, users: users}
	for _, option := range options {
		option(service)
	}
	return service
}
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateProfileRequest) (*ApplicantProfile, error) {
	if principal.EnforceOwner(ctx, userID) != nil {
		return nil, ErrProfileNotFound
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	profileType := req.ProfileType
	if profileType == "" {
		profileType = ProfileTypeMaster
	}
	if !profileType.Valid() {
		return nil, newValidationError("profile_type must be master, domain, or application")
	}
	if profileType != ProfileTypeMaster && req.IsDefault {
		return nil, newValidationError("only a master profile may be the default")
	}
	p := &ApplicantProfile{Base: Base{ID: uuid.New()}, UserID: userID, Name: strings.TrimSpace(req.Name), ProfileType: profileType, ParentProfileID: req.ParentProfileID, Headline: req.Headline, Summary: req.Summary, IsDefault: req.IsDefault}
	if err := s.validateParent(ctx, p); err != nil {
		return nil, err
	}
	existing, err := s.profiles.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		p.IsDefault = true
	}
	if err = s.profiles.Create(ctx, p); err != nil {
		return nil, err
	}
	if p.ProfileType != ProfileTypeMaster && s.workflow != nil {
		profileID, action, entityType := p.ID, "profile.derived.created", "profile"
		_ = s.workflow.CreateAudit(ctx, &AuditLog{UserID: &userID, ProfileID: &profileID, Action: action, EntityType: &entityType, EntityID: &profileID})
	}
	return p, nil
}

func (s *Service) validateParent(ctx context.Context, profile *ApplicantProfile) error {
	if profile.ProfileType == ProfileTypeMaster {
		if profile.ParentProfileID != nil {
			return ErrInvalidProfileParent
		}
		return nil
	}
	if profile.ParentProfileID == nil {
		return ErrProfileParentRequired
	}
	if *profile.ParentProfileID == profile.ID {
		return ErrInvalidProfileParent
	}
	parent, err := s.profiles.GetByID(ctx, *profile.ParentProfileID)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return ErrInvalidProfileParent
		}
		return err
	}
	if parent.UserID != profile.UserID {
		return ErrInvalidProfileParent
	}
	if profile.ProfileType == ProfileTypeDomain && parent.ProfileType != ProfileTypeMaster {
		return ErrInvalidProfileParent
	}
	if profile.ProfileType == ProfileTypeApplication && parent.ProfileType != ProfileTypeDomain && parent.ProfileType != ProfileTypeMaster {
		return ErrInvalidProfileParent
	}
	seen := map[uuid.UUID]bool{profile.ID: true}
	current := parent
	for depth := 0; current != nil; depth++ {
		if depth >= 3 || seen[current.ID] {
			return ErrProfileInheritanceCycle
		}
		seen[current.ID] = true
		if current.ParentProfileID == nil {
			break
		}
		current, err = s.profiles.GetByID(ctx, *current.ParentProfileID)
		if err != nil {
			return ErrInvalidProfileParent
		}
	}
	return nil
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	value, err := s.profiles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, value.UserID) != nil {
		return nil, ErrProfileNotFound
	}
	return value, nil
}
func (s *Service) GetFull(ctx context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.profiles.GetFull(ctx, id)
}
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]ApplicantProfile, error) {
	if principal.EnforceOwner(ctx, userID) != nil {
		return nil, ErrProfileNotFound
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.profiles.ListByUserID(ctx, userID)
}
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateProfileRequest) (*ApplicantProfile, error) {
	p, err := s.profiles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, p.UserID) != nil {
		return nil, ErrProfileNotFound
	}
	if req.Name != nil {
		p.Name = strings.TrimSpace(*req.Name)
	}
	if req.Headline != nil {
		p.Headline = req.Headline
	}
	if req.Summary != nil {
		p.Summary = req.Summary
	}
	if req.IsDefault != nil {
		if *req.IsDefault && p.ProfileType != ProfileTypeMaster {
			return nil, newValidationError("only a master profile may be the default")
		}
		p.IsDefault = *req.IsDefault
	}
	if err = s.profiles.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetPersonalInfo(ctx context.Context, pid uuid.UUID) (*ProfilePersonalInfo, error) {
	if _, err := s.profiles.GetByID(ctx, pid); err != nil {
		return nil, err
	}
	return s.profiles.GetPersonalInfo(ctx, pid)
}
func (s *Service) PutPersonalInfo(ctx context.Context, pid uuid.UUID, req PersonalInfoRequest) (*ProfilePersonalInfo, error) {
	if _, err := s.profiles.GetByID(ctx, pid); err != nil {
		return nil, err
	}
	if req.FirstName == nil || strings.TrimSpace(*req.FirstName) == "" || req.LastName == nil || strings.TrimSpace(*req.LastName) == "" {
		return nil, newValidationError("firstName and lastName are required")
	}
	v := &ProfilePersonalInfo{ProfileID: pid}
	applyPersonal(v, req)
	if err := s.profiles.UpsertPersonalInfo(ctx, v); err != nil {
		return nil, err
	}
	return s.profiles.GetPersonalInfo(ctx, pid)
}
func (s *Service) PatchPersonalInfo(ctx context.Context, pid uuid.UUID, req PersonalInfoRequest) (*ProfilePersonalInfo, error) {
	v, err := s.GetPersonalInfo(ctx, pid)
	if err != nil {
		return nil, err
	}
	applyPersonal(v, req)
	if strings.TrimSpace(v.FirstName) == "" || strings.TrimSpace(v.LastName) == "" {
		return nil, newValidationError("firstName and lastName are required")
	}
	if err = s.profiles.UpsertPersonalInfo(ctx, v); err != nil {
		return nil, err
	}
	return s.profiles.GetPersonalInfo(ctx, pid)
}
func applyPersonal(v *ProfilePersonalInfo, r PersonalInfoRequest) {
	if r.FirstName != nil {
		v.FirstName = strings.TrimSpace(*r.FirstName)
	}
	if r.MiddleName != nil {
		v.MiddleName = r.MiddleName
	}
	if r.LastName != nil {
		v.LastName = strings.TrimSpace(*r.LastName)
	}
	if r.PreferredName != nil {
		v.PreferredName = r.PreferredName
	}
	if r.Phone != nil {
		v.Phone = r.Phone
	}
	if r.City != nil {
		v.City = r.City
	}
	if r.StateOrRegion != nil {
		v.StateOrRegion = r.StateOrRegion
	}
	if r.Country != nil {
		v.Country = r.Country
	}
	if r.Nationality != nil {
		v.Nationality = r.Nationality
	}
	if r.LinkedInURL != nil {
		v.LinkedInURL = r.LinkedInURL
	}
	if r.GitHubURL != nil {
		v.GitHubURL = r.GitHubURL
	}
	if r.WebsiteURL != nil {
		v.WebsiteURL = r.WebsiteURL
	}
}

func (s *Service) ListSection(ctx context.Context, k SectionKind, pid uuid.UUID) (any, error) {
	if _, err := s.profiles.GetByID(ctx, pid); err != nil {
		return nil, err
	}
	switch k {
	case EducationKind:
		return s.sections.ListEducation(ctx, pid)
	case EmploymentKind:
		return s.sections.ListEmployment(ctx, pid)
	case ProjectsKind:
		return s.sections.ListProjects(ctx, pid)
	case PublicationsKind:
		return s.sections.ListPublications(ctx, pid)
	case ArticlesKind:
		return s.sections.ListArticles(ctx, pid)
	case SkillsKind:
		return s.sections.ListSkills(ctx, pid)
	case ResearchInterestsKind:
		return s.sections.ListResearchInterests(ctx, pid)
	case CareerGoalsKind:
		return s.sections.ListCareerGoals(ctx, pid)
	case CertificationsKind:
		return s.sections.ListCertifications(ctx, pid)
	case AwardsKind:
		return s.sections.ListAwards(ctx, pid)
	case VolunteeringKind:
		return s.sections.ListVolunteering(ctx, pid)
	}
	return nil, fmt.Errorf("unknown section %q", k)
}
func (s *Service) GetSection(ctx context.Context, k SectionKind, pid, id uuid.UUID) (any, error) {
	if _, err := s.profiles.GetByID(ctx, pid); err != nil {
		return nil, err
	}
	switch k {
	case EducationKind:
		return s.sections.GetEducation(ctx, pid, id)
	case EmploymentKind:
		return s.sections.GetEmployment(ctx, pid, id)
	case ProjectsKind:
		return s.sections.GetProject(ctx, pid, id)
	case PublicationsKind:
		return s.sections.GetPublication(ctx, pid, id)
	case ArticlesKind:
		return s.sections.GetArticle(ctx, pid, id)
	case SkillsKind:
		return s.sections.GetSkill(ctx, pid, id)
	case ResearchInterestsKind:
		return s.sections.GetResearchInterest(ctx, pid, id)
	case CareerGoalsKind:
		return s.sections.GetCareerGoal(ctx, pid, id)
	case CertificationsKind:
		return s.sections.GetCertification(ctx, pid, id)
	case AwardsKind:
		return s.sections.GetAward(ctx, pid, id)
	case VolunteeringKind:
		return s.sections.GetVolunteer(ctx, pid, id)
	}
	return nil, fmt.Errorf("unknown section %q", k)
}
func (s *Service) CreateSection(ctx context.Context, k SectionKind, pid uuid.UUID, r SectionRequest) (any, error) {
	if _, err := s.profiles.GetByID(ctx, pid); err != nil {
		return nil, err
	}
	if err := validateSection(k, r, true); err != nil {
		return nil, newValidationError(err.Error())
	}
	v := newSection(k, pid, r)
	if err := s.persistCreate(ctx, k, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *Service) UpdateSection(ctx context.Context, k SectionKind, pid, id uuid.UUID, r SectionRequest) (any, error) {
	v, err := s.GetSection(ctx, k, pid, id)
	if err != nil {
		return nil, err
	}
	if err = validateSection(k, r, false); err != nil {
		return nil, newValidationError(err.Error())
	}
	applySection(v, r)
	if err = s.persistUpdate(ctx, k, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *Service) DeleteSection(ctx context.Context, k SectionKind, pid, id uuid.UUID) error {
	if _, err := s.profiles.GetByID(ctx, pid); err != nil {
		return err
	}
	switch k {
	case EducationKind:
		return s.sections.DeleteEducation(ctx, pid, id)
	case EmploymentKind:
		return s.sections.DeleteEmployment(ctx, pid, id)
	case ProjectsKind:
		return s.sections.DeleteProject(ctx, pid, id)
	case PublicationsKind:
		return s.sections.DeletePublication(ctx, pid, id)
	case ArticlesKind:
		return s.sections.DeleteArticle(ctx, pid, id)
	case SkillsKind:
		return s.sections.DeleteSkill(ctx, pid, id)
	case ResearchInterestsKind:
		return s.sections.DeleteResearchInterest(ctx, pid, id)
	case CareerGoalsKind:
		return s.sections.DeleteCareerGoal(ctx, pid, id)
	case CertificationsKind:
		return s.sections.DeleteCertification(ctx, pid, id)
	case AwardsKind:
		return s.sections.DeleteAward(ctx, pid, id)
	case VolunteeringKind:
		return s.sections.DeleteVolunteer(ctx, pid, id)
	}
	return fmt.Errorf("unknown section %q", k)
}
func validateSection(k SectionKind, r SectionRequest, create bool) error {
	required := func(v *string, name string) error {
		if create && (v == nil || strings.TrimSpace(*v) == "") {
			return fmt.Errorf("%s is required", name)
		}
		if v != nil && strings.TrimSpace(*v) == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}
		return nil
	}
	var checks []error
	switch k {
	case EducationKind:
		checks = []error{required(r.Institution, "institution"), required(r.Degree, "degree"), required(r.FieldOfStudy, "fieldOfStudy")}
	case EmploymentKind:
		checks = []error{required(r.Organization, "organization"), required(r.JobTitle, "jobTitle")}
	case ProjectsKind, PublicationsKind, ArticlesKind, AwardsKind:
		checks = []error{required(r.Title, "title")}
	case SkillsKind, ResearchInterestsKind, CertificationsKind:
		checks = []error{required(r.Name, "name")}
	case CareerGoalsKind:
		checks = []error{required(r.Title, "title"), required(r.Description, "description")}
	case VolunteeringKind:
		checks = []error{required(r.Organization, "organization"), required(r.Role, "role")}
	}
	for _, err := range checks {
		if err != nil {
			return err
		}
	}
	if r.StartDate != nil && r.EndDate != nil && r.EndDate.Before(*r.StartDate) {
		return errors.New("endDate cannot be before startDate")
	}
	return nil
}

func (s *Service) persistCreate(c context.Context, k SectionKind, v any) error {
	switch x := v.(type) {
	case *EducationHistory:
		return s.sections.CreateEducation(c, x)
	case *EmploymentHistory:
		return s.sections.CreateEmployment(c, x)
	case *Project:
		return s.sections.CreateProject(c, x)
	case *Publication:
		return s.sections.CreatePublication(c, x)
	case *Article:
		return s.sections.CreateArticle(c, x)
	case *Skill:
		return s.sections.CreateSkill(c, x)
	case *ResearchInterest:
		return s.sections.CreateResearchInterest(c, x)
	case *CareerGoal:
		return s.sections.CreateCareerGoal(c, x)
	case *Certification:
		return s.sections.CreateCertification(c, x)
	case *Award:
		return s.sections.CreateAward(c, x)
	case *VolunteerExperience:
		return s.sections.CreateVolunteer(c, x)
	}
	return fmt.Errorf("unsupported section %q", k)
}
func (s *Service) persistUpdate(c context.Context, k SectionKind, v any) error {
	switch x := v.(type) {
	case *EducationHistory:
		return s.sections.UpdateEducation(c, x)
	case *EmploymentHistory:
		return s.sections.UpdateEmployment(c, x)
	case *Project:
		return s.sections.UpdateProject(c, x)
	case *Publication:
		return s.sections.UpdatePublication(c, x)
	case *Article:
		return s.sections.UpdateArticle(c, x)
	case *Skill:
		return s.sections.UpdateSkill(c, x)
	case *ResearchInterest:
		return s.sections.UpdateResearchInterest(c, x)
	case *CareerGoal:
		return s.sections.UpdateCareerGoal(c, x)
	case *Certification:
		return s.sections.UpdateCertification(c, x)
	case *Award:
		return s.sections.UpdateAward(c, x)
	case *VolunteerExperience:
		return s.sections.UpdateVolunteer(c, x)
	}
	return fmt.Errorf("unsupported section %q", k)
}
