package application

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateApplicationRequest struct {
	Name               string     `json:"name" validate:"required,max=250"`
	ParentProfileID    uuid.UUID  `json:"parentProfileId" validate:"required"`
	InstitutionID      *uuid.UUID `json:"institutionId"`
	ProgrammeID        *uuid.UUID `json:"programmeId"`
	ScholarshipID      *uuid.UUID `json:"scholarshipId"`
	Intake             *string    `json:"intake"`
	IntakeYear         *int       `json:"intakeYear" validate:"omitempty,gte=2000,lte=2200"`
	Country            *string    `json:"country"`
	Priority           *int       `json:"priority" validate:"omitempty,gte=1,lte=5"`
	Notes              *string    `json:"notes"`
	CreateDefaultTasks *bool      `json:"createDefaultTasks"`
}
type UpdateApplicationRequest struct {
	Name                 *string    `json:"name" validate:"omitempty,min=1,max=250"`
	InstitutionID        *uuid.UUID `json:"institutionId"`
	ProgrammeID          *uuid.UUID `json:"programmeId"`
	ScholarshipID        *uuid.UUID `json:"scholarshipId"`
	Intake               *string    `json:"intake"`
	IntakeYear           *int       `json:"intakeYear" validate:"omitempty,gte=2000,lte=2200"`
	Country              *string    `json:"country"`
	Status               *Status    `json:"status"`
	Priority             *int       `json:"priority" validate:"omitempty,gte=1,lte=5"`
	Notes                *string    `json:"notes"`
	ManualStatusOverride bool       `json:"manualStatusOverride"`
}

type NamedResource struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
type ApplicationResponse struct {
	ID                 uuid.UUID      `json:"id"`
	UserID             uuid.UUID      `json:"userId"`
	ApplicantProfileID uuid.UUID      `json:"applicantProfileId"`
	Institution        *NamedResource `json:"institution,omitempty"`
	Programme          *NamedResource `json:"programme,omitempty"`
	Scholarship        *NamedResource `json:"scholarship,omitempty"`
	Name               string         `json:"name"`
	Intake             *string        `json:"intake,omitempty"`
	IntakeYear         *int           `json:"intakeYear,omitempty"`
	Country            *string        `json:"country,omitempty"`
	Status             Status         `json:"status"`
	Priority           *int           `json:"priority,omitempty"`
	Notes              *string        `json:"notes,omitempty"`
	ResearchStatus     ResearchStatus `json:"researchStatus"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

func toApplication(a *Application) ApplicationResponse {
	v := ApplicationResponse{ID: a.ID, UserID: a.UserID, ApplicantProfileID: a.ApplicantProfileID, Name: a.Name, Intake: a.Intake, IntakeYear: a.IntakeYear, Country: a.Country, Status: a.Status, Priority: a.Priority, Notes: a.Notes, ResearchStatus: a.ResearchStatus, CreatedAt: a.CreatedAt.UTC(), UpdatedAt: a.UpdatedAt.UTC()}
	if a.Institution != nil {
		v.Institution = &NamedResource{a.Institution.ID, a.Institution.Name}
	}
	if a.Programme != nil {
		v.Programme = &NamedResource{a.Programme.ID, a.Programme.Name}
	}
	if a.Scholarship != nil {
		v.Scholarship = &NamedResource{a.Scholarship.ID, a.Scholarship.Name}
	}
	return v
}

type RequirementRequest struct {
	Category         string     `json:"category" validate:"required"`
	Title            string     `json:"title" validate:"required"`
	Description      *string    `json:"description"`
	IsMandatory      *bool      `json:"isMandatory"`
	Status           *string    `json:"status" validate:"omitempty,oneof=unknown not_started in_progress satisfied not_satisfied not_applicable waived"`
	DueDate          *time.Time `json:"dueDate"`
	SourceURL        *string    `json:"sourceUrl" validate:"omitempty,url"`
	EvidenceRequired *bool      `json:"evidenceRequired"`
	Notes            *string    `json:"notes"`
	SortOrder        *int       `json:"sortOrder"`
}
type RequirementResponse struct {
	ID               uuid.UUID  `json:"id"`
	ApplicationID    uuid.UUID  `json:"applicationId"`
	Category         string     `json:"category"`
	Title            string     `json:"title"`
	Description      *string    `json:"description,omitempty"`
	IsMandatory      bool       `json:"isMandatory"`
	Status           string     `json:"status"`
	DueDate          *time.Time `json:"dueDate,omitempty"`
	SourceID         *uuid.UUID `json:"sourceId,omitempty"`
	SourceURL        *string    `json:"sourceUrl,omitempty"`
	EvidenceRequired bool       `json:"evidenceRequired"`
	Notes            *string    `json:"notes,omitempty"`
	SortOrder        *int       `json:"sortOrder,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func toRequirement(v *ApplicationRequirement) RequirementResponse {
	return RequirementResponse{v.ID, v.ApplicationID, v.Category, v.Title, v.Description, v.IsMandatory, v.Status, v.DueDate, v.SourceID, v.SourceURL, v.EvidenceRequired, v.Notes, v.SortOrder, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type DeadlineRequest struct {
	DeadlineType    string     `json:"deadlineType" validate:"required"`
	Title           string     `json:"title" validate:"required"`
	DeadlineAt      *time.Time `json:"deadlineAt"`
	Timezone        *string    `json:"timezone"`
	DatePrecision   string     `json:"datePrecision" validate:"required,oneof=exact day month approximate unknown"`
	RawDeadlineText *string    `json:"rawDeadlineText"`
	IsHardDeadline  *bool      `json:"isHardDeadline"`
	SourceURL       *string    `json:"sourceUrl" validate:"omitempty,url"`
	VerifiedAt      *time.Time `json:"verifiedAt"`
	Notes           *string    `json:"notes"`
}
type DeadlineResponse struct {
	ID              uuid.UUID  `json:"id"`
	ApplicationID   uuid.UUID  `json:"applicationId"`
	DeadlineType    string     `json:"deadlineType"`
	Title           string     `json:"title"`
	DeadlineAt      *time.Time `json:"deadlineAt,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	DatePrecision   string     `json:"datePrecision"`
	RawDeadlineText *string    `json:"rawDeadlineText,omitempty"`
	IsHardDeadline  bool       `json:"isHardDeadline"`
	SourceID        *uuid.UUID `json:"sourceId,omitempty"`
	SourceURL       *string    `json:"sourceUrl,omitempty"`
	VerifiedAt      *time.Time `json:"verifiedAt,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	Urgency         string     `json:"urgency"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func toDeadline(v *ApplicationDeadline, now time.Time) DeadlineResponse {
	return DeadlineResponse{v.ID, v.ApplicationID, v.DeadlineType, v.Title, utcPtr(v.DeadlineAt), v.Timezone, v.DatePrecision, v.RawDeadlineText, v.IsHardDeadline, v.SourceID, v.SourceURL, utcPtr(v.VerifiedAt), v.Notes, DeadlineUrgency(now, v.DeadlineAt, v.DatePrecision), v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type FundingRequest struct {
	FundingType           string   `json:"fundingType" validate:"required"`
	Currency              *string  `json:"currency"`
	Amount                *float64 `json:"amount" validate:"omitempty,gte=0"`
	AmountPeriod          *string  `json:"amountPeriod"`
	TuitionCoverage       *string  `json:"tuitionCoverage"`
	StipendAmount         *float64 `json:"stipendAmount" validate:"omitempty,gte=0"`
	StipendPeriod         *string  `json:"stipendPeriod"`
	TravelCoverage        *string  `json:"travelCoverage"`
	InsuranceCoverage     *string  `json:"insuranceCoverage"`
	AccommodationCoverage *string  `json:"accommodationCoverage"`
	OtherBenefits         *string  `json:"otherBenefits"`
	Conditions            *string  `json:"conditions"`
	SourceURL             *string  `json:"sourceUrl" validate:"omitempty,url"`
}
type FundingResponse struct {
	ID                    uuid.UUID  `json:"id"`
	ApplicationID         uuid.UUID  `json:"applicationId"`
	FundingType           string     `json:"fundingType"`
	Currency              *string    `json:"currency,omitempty"`
	Amount                *float64   `json:"amount,omitempty"`
	AmountPeriod          *string    `json:"amountPeriod,omitempty"`
	TuitionCoverage       *string    `json:"tuitionCoverage,omitempty"`
	StipendAmount         *float64   `json:"stipendAmount,omitempty"`
	StipendPeriod         *string    `json:"stipendPeriod,omitempty"`
	TravelCoverage        *string    `json:"travelCoverage,omitempty"`
	InsuranceCoverage     *string    `json:"insuranceCoverage,omitempty"`
	AccommodationCoverage *string    `json:"accommodationCoverage,omitempty"`
	OtherBenefits         *string    `json:"otherBenefits,omitempty"`
	Conditions            *string    `json:"conditions,omitempty"`
	SourceID              *uuid.UUID `json:"sourceId,omitempty"`
	SourceURL             *string    `json:"sourceUrl,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

func toFunding(v *ApplicationFunding) FundingResponse {
	return FundingResponse{v.ID, v.ApplicationID, v.FundingType, v.Currency, v.Amount, v.AmountPeriod, v.TuitionCoverage, v.StipendAmount, v.StipendPeriod, v.TravelCoverage, v.InsuranceCoverage, v.AccommodationCoverage, v.OtherBenefits, v.Conditions, v.SourceID, v.SourceURL, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type ContactRequest struct {
	Name         *string `json:"name"`
	Role         *string `json:"role"`
	Email        *string `json:"email" validate:"omitempty,email"`
	Phone        *string `json:"phone"`
	Organization *string `json:"organization"`
	ContactType  *string `json:"contactType"`
	URL          *string `json:"url" validate:"omitempty,url"`
	Notes        *string `json:"notes"`
}
type ContactResponse struct {
	ID            uuid.UUID  `json:"id"`
	ApplicationID uuid.UUID  `json:"applicationId"`
	Name          *string    `json:"name,omitempty"`
	Role          *string    `json:"role,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Organization  *string    `json:"organization,omitempty"`
	ContactType   *string    `json:"contactType,omitempty"`
	URL           *string    `json:"url,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	SourceID      *uuid.UUID `json:"sourceId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func toContact(v *ApplicationContact) ContactResponse {
	return ContactResponse{v.ID, v.ApplicationID, v.Name, v.Role, v.Email, v.Phone, v.Organization, v.ContactType, v.URL, v.Notes, v.SourceID, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type SupervisorRequest struct {
	Name          string  `json:"name" validate:"required"`
	Title         *string `json:"title"`
	Department    *string `json:"department"`
	Institution   *string `json:"institution"`
	Email         *string `json:"email" validate:"omitempty,email"`
	ProfileURL    *string `json:"profileUrl" validate:"omitempty,url"`
	ResearchAreas *string `json:"researchAreas"`
	ContactStatus *string `json:"contactStatus" validate:"omitempty,oneof=not_contacted planned contacted replied interested not_available declined"`
	Notes         *string `json:"notes"`
}
type SupervisorResponse struct {
	ID            uuid.UUID  `json:"id"`
	ApplicationID uuid.UUID  `json:"applicationId"`
	Name          string     `json:"name"`
	Title         *string    `json:"title,omitempty"`
	Department    *string    `json:"department,omitempty"`
	Institution   *string    `json:"institution,omitempty"`
	Email         *string    `json:"email,omitempty"`
	ProfileURL    *string    `json:"profileUrl,omitempty"`
	ResearchAreas *string    `json:"researchAreas,omitempty"`
	ContactStatus *string    `json:"contactStatus,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	SourceID      *uuid.UUID `json:"sourceId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func toSupervisor(v *ApplicationSupervisor) SupervisorResponse {
	return SupervisorResponse{v.ID, v.ApplicationID, v.Name, v.Title, v.Department, v.Institution, v.Email, v.ProfileURL, v.ResearchAreas, v.ContactStatus, v.Notes, v.SourceID, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type URLRequest struct {
	URLType    string  `json:"urlType" validate:"required"`
	Label      *string `json:"label"`
	URL        string  `json:"url" validate:"required,url"`
	IsOfficial bool    `json:"isOfficial"`
}
type URLResponse struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"applicationId"`
	URLType       string    `json:"urlType"`
	Label         *string   `json:"label,omitempty"`
	URL           string    `json:"url"`
	IsOfficial    bool      `json:"isOfficial"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func toURL(v *ApplicationURL) URLResponse {
	return URLResponse{v.ID, v.ApplicationID, v.URLType, v.Label, v.URL, v.IsOfficial, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type TaskRequest struct {
	ParentTaskID *uuid.UUID `json:"parentTaskId"`
	Title        string     `json:"title" validate:"required"`
	Description  *string    `json:"description"`
	Status       *string    `json:"status" validate:"omitempty,oneof=todo in_progress blocked done cancelled"`
	Priority     *int       `json:"priority" validate:"omitempty,gte=1,lte=5"`
	DueAt        *time.Time `json:"dueAt"`
	SortOrder    *int       `json:"sortOrder"`
	TaskType     *string    `json:"taskType"`
	Source       *string    `json:"source"`
}
type TaskResponse struct {
	ID            uuid.UUID  `json:"id"`
	ApplicationID uuid.UUID  `json:"applicationId"`
	ParentTaskID  *uuid.UUID `json:"parentTaskId,omitempty"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	Status        string     `json:"status"`
	Priority      *int       `json:"priority,omitempty"`
	DueAt         *time.Time `json:"dueAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	SortOrder     *int       `json:"sortOrder,omitempty"`
	TaskType      *string    `json:"taskType,omitempty"`
	Source        *string    `json:"source,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func toTask(v *ApplicationTask) TaskResponse {
	return TaskResponse{v.ID, v.ApplicationID, v.ParentTaskID, v.Title, v.Description, v.Status, v.Priority, utcPtr(v.DueAt), utcPtr(v.CompletedAt), v.SortOrder, v.TaskType, v.Source, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type EligibilitySummary struct {
	MandatoryTotal int `json:"mandatoryTotal"`
	Satisfied      int `json:"satisfied"`
	NotSatisfied   int `json:"notSatisfied"`
	Unknown        int `json:"unknown"`
}
type ReadinessSummary struct {
	Status         string `json:"status"`
	CompletedItems int    `json:"completedItems"`
	TotalItems     int    `json:"totalItems"`
	Notice         string `json:"notice"`
}
type ProgressSummary struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}
type ApplicationSummaryResponse struct {
	ApplicationResponse
	NextDeadline          *DeadlineResponse `json:"nextDeadline,omitempty"`
	TaskCompletion        ProgressSummary   `json:"taskCompletion"`
	RequirementCompletion ProgressSummary   `json:"requirementCompletion"`
	Readiness             ReadinessSummary  `json:"readiness"`
}
type ApplicationDetailResponse struct {
	Application       ApplicationResponse   `json:"application"`
	Requirements      []RequirementResponse `json:"requirements"`
	Deadlines         []DeadlineResponse    `json:"deadlines"`
	Funding           []FundingResponse     `json:"funding"`
	Contacts          []ContactResponse     `json:"contacts"`
	Supervisors       []SupervisorResponse  `json:"supervisors"`
	URLs              []URLResponse         `json:"urls"`
	Tasks             []TaskResponse        `json:"tasks"`
	LatestResearchRun *ResearchRunResponse  `json:"latestResearchRun,omitempty"`
}

type ResearchRunRequest struct {
	Trigger      string `json:"trigger" validate:"omitempty,oneof=manual import scheduled refresh agent"`
	ResearchType string `json:"researchType" validate:"omitempty,oneof=full deadline_refresh funding_refresh eligibility programme scholarship supervisor other"`
}
type ResearchRunResponse struct {
	ID             uuid.UUID  `json:"id"`
	ApplicationID  *uuid.UUID `json:"applicationId,omitempty"`
	ResearchTaskID *uuid.UUID `json:"researchTaskId,omitempty"`
	Status         string     `json:"status"`
	Trigger        string     `json:"trigger"`
	ResearchType   string     `json:"researchType"`
	ModelProvider  *string    `json:"modelProvider,omitempty"`
	ModelName      *string    `json:"modelName,omitempty"`
	PromptVersion  *string    `json:"promptVersion,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	ErrorMessage   *string    `json:"errorMessage,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func toResearchRun(v *ResearchRun) ResearchRunResponse {
	return ResearchRunResponse{v.ID, v.ApplicationID, v.ResearchTaskID, v.Status, v.Trigger, v.ResearchType, v.ModelProvider, v.ModelName, v.PromptVersion, utcPtr(v.StartedAt), utcPtr(v.CompletedAt), v.ErrorMessage, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}

type ResearchSourceResponse struct {
	ID            uuid.UUID  `json:"id"`
	ResearchRunID uuid.UUID  `json:"researchRunId"`
	ApplicationID *uuid.UUID `json:"applicationId,omitempty"`
	URL           string     `json:"url"`
	Title         *string    `json:"title,omitempty"`
	Publisher     *string    `json:"publisher,omitempty"`
	SourceType    string     `json:"sourceType"`
	IsOfficial    bool       `json:"isOfficial"`
	RetrievedAt   time.Time  `json:"retrievedAt"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	ContentHash   *string    `json:"contentHash,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func toResearchSource(v *ResearchSource) ResearchSourceResponse {
	return ResearchSourceResponse{v.ID, v.ResearchRunID, v.ApplicationID, v.URL, v.Title, v.Publisher, v.SourceType, v.IsOfficial, v.RetrievedAt.UTC(), utcPtr(v.PublishedAt), v.ContentHash, v.Notes, v.CreatedAt.UTC()}
}

type ResearchFindingResponse struct {
	ID                 uuid.UUID       `json:"id"`
	ResearchRunID      uuid.UUID       `json:"researchRunId"`
	ApplicationID      *uuid.UUID      `json:"applicationId,omitempty"`
	SourceID           *uuid.UUID      `json:"sourceId,omitempty"`
	Category           string          `json:"category"`
	Field              string          `json:"field"`
	Value              json.RawMessage `json:"value"`
	Confidence         *float64        `json:"confidence,omitempty"`
	VerificationStatus string          `json:"verificationStatus"`
	ReviewStatus       string          `json:"reviewStatus"`
	RawText            *string         `json:"rawText,omitempty"`
	Notes              *string         `json:"notes,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
}

func toResearchFinding(v *ResearchFinding) ResearchFindingResponse {
	return ResearchFindingResponse{v.ID, v.ResearchRunID, v.ApplicationID, v.SourceID, v.Category, v.Field, json.RawMessage(v.Value), v.Confidence, v.VerificationStatus, v.ReviewStatus, v.RawText, v.Notes, v.CreatedAt.UTC()}
}
func utcPtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	u := v.UTC()
	return &u
}

type ReviewFindingRequest struct {
	ReviewStatus string `json:"reviewStatus" validate:"required,oneof=accepted rejected"`
}

type EvidenceRequest struct {
	EntityType string     `json:"entityType" validate:"required"`
	EntityID   uuid.UUID  `json:"entityId" validate:"required"`
	EvidenceID *uuid.UUID `json:"evidenceId"`
	Status     string     `json:"status" validate:"required,oneof=suggested accepted rejected"`
	Notes      *string    `json:"notes"`
}
type EvidenceResponse struct {
	ID            uuid.UUID  `json:"id"`
	RequirementID uuid.UUID  `json:"requirementId"`
	ProfileID     uuid.UUID  `json:"profileId"`
	EntityType    string     `json:"entityType"`
	EntityID      uuid.UUID  `json:"entityId"`
	EvidenceID    *uuid.UUID `json:"evidenceId,omitempty"`
	Status        string     `json:"status"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}
type EvidenceSuggestion struct {
	EntityType string    `json:"entityType"`
	EntityID   uuid.UUID `json:"entityId"`
	Label      string    `json:"label"`
	Reason     string    `json:"reason"`
}
type DashboardResponse struct {
	ActiveApplications              []ApplicationResponse `json:"activeApplications"`
	UpcomingDeadlines               []DeadlineResponse    `json:"upcomingDeadlines"`
	TasksDueSoon                    []TaskResponse        `json:"tasksDueSoon"`
	ApplicationsNeedingResearch     int                   `json:"applicationsNeedingResearch"`
	ResearchPendingReview           int64                 `json:"researchPendingReview"`
	UnresolvedResearchConflicts     int64                 `json:"unresolvedResearchConflicts"`
	IncompleteMandatoryRequirements int64                 `json:"incompleteMandatoryRequirements"`
	StatusSummary                   map[Status]int        `json:"statusSummary"`
}
