package profile

import (
	"github.com/google/uuid"
	"time"
)

type CreateProfileRequest struct {
	Name      string  `json:"name" validate:"required"`
	Headline  *string `json:"headline"`
	Summary   *string `json:"summary"`
	IsDefault bool    `json:"isDefault"`
}
type UpdateProfileRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1"`
	Headline  *string `json:"headline"`
	Summary   *string `json:"summary"`
	IsDefault *bool   `json:"isDefault"`
}
type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Headline  *string   `json:"headline,omitempty"`
	Summary   *string   `json:"summary,omitempty"`
	IsDefault bool      `json:"isDefault"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func profileResponse(p *ApplicantProfile) ProfileResponse {
	return ProfileResponse{ID: p.ID, UserID: p.UserID, Name: p.Name, Headline: p.Headline, Summary: p.Summary, IsDefault: p.IsDefault, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
}

type PersonalInfoRequest struct {
	FirstName     *string `json:"firstName"`
	MiddleName    *string `json:"middleName"`
	LastName      *string `json:"lastName"`
	PreferredName *string `json:"preferredName"`
	Phone         *string `json:"phone"`
	City          *string `json:"city"`
	StateOrRegion *string `json:"stateOrRegion"`
	Country       *string `json:"country"`
	Nationality   *string `json:"nationality"`
	LinkedInURL   *string `json:"linkedInUrl" validate:"omitempty,url"`
	GitHubURL     *string `json:"githubUrl" validate:"omitempty,url"`
	WebsiteURL    *string `json:"websiteUrl" validate:"omitempty,url"`
}
type PersonalInfoResponse struct {
	ID            uuid.UUID `json:"id"`
	ProfileID     uuid.UUID `json:"profileId"`
	FirstName     string    `json:"firstName"`
	MiddleName    *string   `json:"middleName,omitempty"`
	LastName      string    `json:"lastName"`
	PreferredName *string   `json:"preferredName,omitempty"`
	Phone         *string   `json:"phone,omitempty"`
	City          *string   `json:"city,omitempty"`
	StateOrRegion *string   `json:"stateOrRegion,omitempty"`
	Country       *string   `json:"country,omitempty"`
	Nationality   *string   `json:"nationality,omitempty"`
	LinkedInURL   *string   `json:"linkedInUrl,omitempty"`
	GitHubURL     *string   `json:"githubUrl,omitempty"`
	WebsiteURL    *string   `json:"websiteUrl,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// SectionRequest is the transport DTO shared by the structurally similar list sections.
// Each route validates only the fields meaningful to that section.
type SectionRequest struct {
	Institution     *string    `json:"institution"`
	Degree          *string    `json:"degree"`
	FieldOfStudy    *string    `json:"fieldOfStudy"`
	Organization    *string    `json:"organization"`
	JobTitle        *string    `json:"jobTitle"`
	EmploymentType  *string    `json:"employmentType"`
	Title           *string    `json:"title"`
	Name            *string    `json:"name"`
	Role            *string    `json:"role"`
	PublicationType *string    `json:"publicationType"`
	Publisher       *string    `json:"publisher"`
	PublicationDate *time.Time `json:"publicationDate"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
	IsCurrent       *bool      `json:"isCurrent"`
	Grade           *string    `json:"grade"`
	GradeScale      *string    `json:"gradeScale"`
	Classification  *string    `json:"classification"`
	City            *string    `json:"city"`
	Country         *string    `json:"country"`
	Description     *string    `json:"description"`
	URL             *string    `json:"url" validate:"omitempty,url"`
	RepositoryURL   *string    `json:"repositoryUrl" validate:"omitempty,url"`
	DOI             *string    `json:"doi"`
	Category        *string    `json:"category"`
	Proficiency     *string    `json:"proficiency"`
	YearsExperience *float64   `json:"yearsExperience" validate:"omitempty,gte=0"`
	Priority        *int       `json:"priority" validate:"omitempty,gte=0"`
	GoalType        *string    `json:"goalType"`
	Issuer          *string    `json:"issuer"`
	IssueDate       *time.Time `json:"issueDate"`
	ExpirationDate  *time.Time `json:"expirationDate"`
	CredentialID    *string    `json:"credentialId"`
	AwardDate       *time.Time `json:"awardDate"`
}
type SectionResponse struct {
	ID              uuid.UUID  `json:"id"`
	ProfileID       uuid.UUID  `json:"profileId"`
	Institution     *string    `json:"institution,omitempty"`
	Degree          *string    `json:"degree,omitempty"`
	FieldOfStudy    *string    `json:"fieldOfStudy,omitempty"`
	Organization    *string    `json:"organization,omitempty"`
	JobTitle        *string    `json:"jobTitle,omitempty"`
	EmploymentType  *string    `json:"employmentType,omitempty"`
	Title           *string    `json:"title,omitempty"`
	Name            *string    `json:"name,omitempty"`
	Role            *string    `json:"role,omitempty"`
	PublicationType *string    `json:"publicationType,omitempty"`
	Publisher       *string    `json:"publisher,omitempty"`
	PublicationDate *time.Time `json:"publicationDate,omitempty"`
	StartDate       *time.Time `json:"startDate,omitempty"`
	EndDate         *time.Time `json:"endDate,omitempty"`
	IsCurrent       *bool      `json:"isCurrent,omitempty"`
	Grade           *string    `json:"grade,omitempty"`
	GradeScale      *string    `json:"gradeScale,omitempty"`
	Classification  *string    `json:"classification,omitempty"`
	City            *string    `json:"city,omitempty"`
	Country         *string    `json:"country,omitempty"`
	Description     *string    `json:"description,omitempty"`
	URL             *string    `json:"url,omitempty"`
	RepositoryURL   *string    `json:"repositoryUrl,omitempty"`
	DOI             *string    `json:"doi,omitempty"`
	Category        *string    `json:"category,omitempty"`
	Proficiency     *string    `json:"proficiency,omitempty"`
	YearsExperience *float64   `json:"yearsExperience,omitempty"`
	Priority        *int       `json:"priority,omitempty"`
	GoalType        *string    `json:"goalType,omitempty"`
	Issuer          *string    `json:"issuer,omitempty"`
	IssueDate       *time.Time `json:"issueDate,omitempty"`
	ExpirationDate  *time.Time `json:"expirationDate,omitempty"`
	CredentialID    *string    `json:"credentialId,omitempty"`
	AwardDate       *time.Time `json:"awardDate,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}
type FullProfileResponse struct {
	ProfileResponse
	PersonalInfo      *PersonalInfoResponse `json:"personalInfo"`
	Education         []SectionResponse     `json:"education"`
	Employment        []SectionResponse     `json:"employment"`
	Projects          []SectionResponse     `json:"projects"`
	Publications      []SectionResponse     `json:"publications"`
	Articles          []SectionResponse     `json:"articles"`
	Skills            []SectionResponse     `json:"skills"`
	ResearchInterests []SectionResponse     `json:"researchInterests"`
	CareerGoals       []SectionResponse     `json:"careerGoals"`
	Certifications    []SectionResponse     `json:"certifications"`
	Awards            []SectionResponse     `json:"awards"`
	Volunteering      []SectionResponse     `json:"volunteering"`
}

// Named aliases keep generated API documentation and call sites domain-specific.
type CreateEducationRequest = SectionRequest
type UpdateEducationRequest = SectionRequest
type EducationResponse = SectionResponse
type CreateEmploymentRequest = SectionRequest
type UpdateEmploymentRequest = SectionRequest
type EmploymentResponse = SectionResponse
type CreateProjectRequest = SectionRequest
type UpdateProjectRequest = SectionRequest
type ProjectResponse = SectionResponse
type CreatePublicationRequest = SectionRequest
type UpdatePublicationRequest = SectionRequest
type PublicationResponse = SectionResponse
type CreateArticleRequest = SectionRequest
type UpdateArticleRequest = SectionRequest
type ArticleResponse = SectionResponse
type CreateSkillRequest = SectionRequest
type UpdateSkillRequest = SectionRequest
type SkillResponse = SectionResponse
type CreateResearchInterestRequest = SectionRequest
type UpdateResearchInterestRequest = SectionRequest
type ResearchInterestResponse = SectionResponse
type CreateCareerGoalRequest = SectionRequest
type UpdateCareerGoalRequest = SectionRequest
type CareerGoalResponse = SectionResponse
type CreateCertificationRequest = SectionRequest
type UpdateCertificationRequest = SectionRequest
type CertificationResponse = SectionResponse
type CreateAwardRequest = SectionRequest
type UpdateAwardRequest = SectionRequest
type AwardResponse = SectionResponse
type CreateVolunteerRequest = SectionRequest
type UpdateVolunteerRequest = SectionRequest
type VolunteerResponse = SectionResponse
