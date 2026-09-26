package profile

import (
	"context"
	"errors"
	"strings"
	"time"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

func (s *Service) ExportBundle(ctx context.Context, profileID uuid.UUID) (*ProfileBundle, error) {
	if _, err := s.Get(ctx, profileID); err != nil {
		return nil, err
	}
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	data := ProfileBundleData{
		Name:              effective.Name,
		ProfileType:       effective.ProfileType,
		ParentProfileID:   effective.ParentProfileID,
		Headline:          effective.Headline,
		Summary:           effective.Summary,
		IsDefault:         effective.IsDefault,
		Education:         sectionRequests(effective.Education),
		Employment:        sectionRequests(effective.Employment),
		Projects:          sectionRequests(effective.Projects),
		Publications:      sectionRequests(effective.Publications),
		Articles:          sectionRequests(effective.Articles),
		Skills:            sectionRequests(effective.Skills),
		ResearchInterests: sectionRequests(effective.ResearchInterests),
		CareerGoals:       sectionRequests(effective.CareerGoals),
		Certifications:    sectionRequests(effective.Certifications),
		Awards:            sectionRequests(effective.Awards),
		Volunteering:      sectionRequests(effective.Volunteering),
	}
	if effective.PersonalInfo != nil {
		data.PersonalInfo = personalRequest(effective.PersonalInfo)
	}
	return &ProfileBundle{SchemaVersion: ProfileBundleSchemaVersion, ExportedAt: time.Now().UTC(), Profile: data}, nil
}

func (s *Service) ImportBundle(ctx context.Context, userID uuid.UUID, request ImportProfileBundleRequest) (*ApplicantProfile, error) {
	if principal.EnforceOwner(ctx, userID) != nil {
		return nil, ErrProfileNotFound
	}
	if request.SchemaVersion != ProfileBundleSchemaVersion {
		return nil, newValidationError("schemaVersion must be " + ProfileBundleSchemaVersion)
	}
	if s.bundles == nil {
		return nil, errors.New("profile bundle repository is not configured")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	data := request.Profile
	if !data.ProfileType.Valid() {
		return nil, newValidationError("profileType must be master, domain, or application")
	}
	if data.ProfileType != ProfileTypeMaster && data.IsDefault {
		return nil, newValidationError("only a master profile may be the default")
	}
	profile := &ApplicantProfile{
		Base:            Base{ID: uuid.New()},
		UserID:          userID,
		Name:            strings.TrimSpace(data.Name),
		ProfileType:     data.ProfileType,
		ParentProfileID: data.ParentProfileID,
		Headline:        data.Headline,
		Summary:         data.Summary,
		IsDefault:       data.IsDefault,
	}
	if profile.Name == "" {
		return nil, newValidationError("profile name is required")
	}
	if err := s.validateParent(ctx, profile); err != nil {
		return nil, err
	}
	existing, err := s.profiles.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		profile.IsDefault = true
	}

	var personal *ProfilePersonalInfo
	if data.PersonalInfo != nil {
		if data.PersonalInfo.FirstName == nil || strings.TrimSpace(*data.PersonalInfo.FirstName) == "" || data.PersonalInfo.LastName == nil || strings.TrimSpace(*data.PersonalInfo.LastName) == "" {
			return nil, newValidationError("personalInfo.firstName and personalInfo.lastName are required")
		}
		personal = &ProfilePersonalInfo{ProfileID: profile.ID}
		applyPersonal(personal, *data.PersonalInfo)
	}

	sections := make([]any, 0)
	groups := []struct {
		kind SectionKind
		rows []SectionRequest
	}{
		{EducationKind, data.Education},
		{EmploymentKind, data.Employment},
		{ProjectsKind, data.Projects},
		{PublicationsKind, data.Publications},
		{ArticlesKind, data.Articles},
		{SkillsKind, data.Skills},
		{ResearchInterestsKind, data.ResearchInterests},
		{CareerGoalsKind, data.CareerGoals},
		{CertificationsKind, data.Certifications},
		{AwardsKind, data.Awards},
		{VolunteeringKind, data.Volunteering},
	}
	for _, group := range groups {
		for _, row := range group.rows {
			if err := validateSection(group.kind, row, true); err != nil {
				return nil, newValidationError(string(group.kind) + ": " + err.Error())
			}
			sections = append(sections, newSection(group.kind, profile.ID, row))
		}
	}

	entityType := "profile"
	audit := &AuditLog{UserID: &userID, ProfileID: &profile.ID, Action: "profile.bundle.imported", EntityType: &entityType, EntityID: &profile.ID}
	if err := s.bundles.CreateBundle(ctx, profile, personal, sections, audit); err != nil {
		return nil, err
	}
	return s.profiles.GetFull(ctx, profile.ID)
}

func personalRequest(value *PersonalInfoResponse) *PersonalInfoRequest {
	firstName, lastName := value.FirstName, value.LastName
	return &PersonalInfoRequest{
		FirstName:     &firstName,
		MiddleName:    value.MiddleName,
		LastName:      &lastName,
		PreferredName: value.PreferredName,
		Phone:         value.Phone,
		City:          value.City,
		StateOrRegion: value.StateOrRegion,
		Country:       value.Country,
		Nationality:   value.Nationality,
		LinkedInURL:   value.LinkedInURL,
		GitHubURL:     value.GitHubURL,
		WebsiteURL:    value.WebsiteURL,
	}
}

func sectionRequests(values []SectionResponse) []SectionRequest {
	result := make([]SectionRequest, 0, len(values))
	for _, value := range values {
		result = append(result, SectionRequest{
			Institution: value.Institution, Degree: value.Degree, FieldOfStudy: value.FieldOfStudy,
			Organization: value.Organization, JobTitle: value.JobTitle, EmploymentType: value.EmploymentType,
			Title: value.Title, Name: value.Name, Role: value.Role, PublicationType: value.PublicationType,
			Publisher: value.Publisher, PublicationDate: value.PublicationDate, StartDate: value.StartDate,
			EndDate: value.EndDate, IsCurrent: value.IsCurrent, Grade: value.Grade, GradeScale: value.GradeScale,
			Classification: value.Classification, City: value.City, Country: value.Country,
			Description: value.Description, URL: value.URL, RepositoryURL: value.RepositoryURL, DOI: value.DOI,
			Category: value.Category, Proficiency: value.Proficiency, YearsExperience: value.YearsExperience,
			Priority: value.Priority, GoalType: value.GoalType, Issuer: value.Issuer, IssueDate: value.IssueDate,
			ExpirationDate: value.ExpirationDate, CredentialID: value.CredentialID, AwardDate: value.AwardDate,
		})
	}
	return result
}
