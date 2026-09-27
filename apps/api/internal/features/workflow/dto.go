package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateResearchTaskRequest struct {
	UserID              uuid.UUID                   `json:"userId" validate:"required"`
	ParentTaskID        *uuid.UUID                  `json:"parentTaskId"`
	TargetApplicationID *uuid.UUID                  `json:"targetApplicationId"`
	ProfileID           *uuid.UUID                  `json:"profileId"`
	Title               string                      `json:"title" validate:"required,max=250"`
	Description         *string                     `json:"description"`
	Instructions        *string                     `json:"instructions"`
	TaskType            string                      `json:"taskType" validate:"required"`
	Priority            *string                     `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
	DueAt               *time.Time                  `json:"dueAt"`
	ResearchConfig      json.RawMessage             `json:"researchConfig"`
	Links               []TaskLinkRequest           `json:"links" validate:"dive"`
	ResearchContextIDs  []uuid.UUID                 `json:"researchContextIds" validate:"dive,required"`
	NewResearchContexts []NewResearchContextRequest `json:"newResearchContexts" validate:"dive"`
}

type NewResearchContextRequest struct {
	Question string `json:"question" validate:"required,max=2000"`
	Answer   string `json:"answer" validate:"required,max=20000"`
}

type CreateResearchContextRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	NewResearchContextRequest
}

type AttachResearchContextsRequest struct {
	ResearchContextIDs []uuid.UUID `json:"researchContextIds" validate:"required,min=1,dive,required"`
}

type ResearchContextResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateResearchTaskRequest struct {
	ParentTaskID        *uuid.UUID      `json:"parentTaskId"`
	TargetApplicationID *uuid.UUID      `json:"targetApplicationId"`
	ProfileID           *uuid.UUID      `json:"profileId"`
	ClearProfile        bool            `json:"clearProfile"`
	Title               *string         `json:"title" validate:"omitempty,min=1,max=250"`
	Description         *string         `json:"description"`
	Instructions        *string         `json:"instructions"`
	TaskType            *string         `json:"taskType"`
	Priority            *string         `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
	DueAt               *time.Time      `json:"dueAt"`
	ResearchConfig      json.RawMessage `json:"researchConfig"`
}

type ResearchTaskFilters struct {
	UserID      *uuid.UUID
	Status      *string
	TaskType    *string
	Priority    *string
	Limit       int
	OldestFirst bool
}

type CompleteAgentResearchRequest struct {
	ResearchRunID uuid.UUID `json:"researchRunId" validate:"required"`
}

type ClaimedResearchTaskResponse struct {
	Task        ResearchTaskResponse `json:"task"`
	ResearchRun ResearchRunDTO       `json:"researchRun"`
}

type TaskLinkRequest struct {
	Label    *string `json:"label"`
	URL      string  `json:"url" validate:"required,url"`
	LinkType *string `json:"linkType"`
}

type ResearchTaskResponse struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"userId"`
	ParentTaskID        *uuid.UUID      `json:"parentTaskId,omitempty"`
	TargetApplicationID *uuid.UUID      `json:"targetApplicationId,omitempty"`
	ProfileID           *uuid.UUID      `json:"profileId,omitempty"`
	Title               string          `json:"title"`
	Description         *string         `json:"description,omitempty"`
	Instructions        *string         `json:"instructions,omitempty"`
	TaskType            string          `json:"taskType"`
	Status              string          `json:"status"`
	Priority            *string         `json:"priority,omitempty"`
	DueAt               *time.Time      `json:"dueAt,omitempty"`
	ResearchConfig      json.RawMessage `json:"researchConfig,omitempty"`
	StartedAt           *time.Time      `json:"startedAt,omitempty"`
	CompletedAt         *time.Time      `json:"completedAt,omitempty"`
	FailedAt            *time.Time      `json:"failedAt,omitempty"`
	FailureReason       *string         `json:"failureReason,omitempty"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type TaskLinkResponse struct {
	ID             uuid.UUID `json:"id"`
	ResearchTaskID uuid.UUID `json:"researchTaskId"`
	Label          *string   `json:"label,omitempty"`
	URL            string    `json:"url"`
	LinkType       *string   `json:"linkType,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type TaskOutputResponse struct {
	ID             uuid.UUID `json:"id"`
	ResearchTaskID uuid.UUID `json:"researchTaskId"`
	OutputType     string    `json:"outputType"`
	EntityID       uuid.UUID `json:"entityId"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AgentResearchSourceRequest struct {
	ResearchRunID uuid.UUID  `json:"researchRunId" validate:"required"`
	URL           string     `json:"url" validate:"required,url"`
	Title         *string    `json:"title"`
	Publisher     *string    `json:"publisher"`
	SourceType    string     `json:"sourceType" validate:"required"`
	IsOfficial    bool       `json:"isOfficial"`
	RetrievedAt   *time.Time `json:"retrievedAt"`
	PublishedAt   *time.Time `json:"publishedAt"`
	ContentHash   *string    `json:"contentHash"`
	Notes         *string    `json:"notes"`
}

type AgentResearchFindingRequest struct {
	ResearchRunID      uuid.UUID       `json:"researchRunId" validate:"required"`
	SourceID           *uuid.UUID      `json:"sourceId"`
	Category           string          `json:"category" validate:"required"`
	Field              string          `json:"field" validate:"required"`
	Value              json.RawMessage `json:"value" validate:"required"`
	Confidence         *float64        `json:"confidence" validate:"omitempty,gte=0,lte=1"`
	VerificationStatus string          `json:"verificationStatus" validate:"required,oneof=verified supported conflicting unverified inferred"`
	RawText            *string         `json:"rawText"`
	Notes              *string         `json:"notes"`
}

type AgentApplicationProposalRequest struct {
	ResearchTaskID      uuid.UUID       `json:"researchTaskId" validate:"required"`
	ResearchRunID       uuid.UUID       `json:"researchRunId" validate:"required"`
	InstitutionID       *uuid.UUID      `json:"institutionId"`
	ProgrammeID         *uuid.UUID      `json:"programmeId"`
	ScholarshipID       *uuid.UUID      `json:"scholarshipId"`
	ProposedInstitution json.RawMessage `json:"proposedInstitution"`
	ProposedProgramme   json.RawMessage `json:"proposedProgramme"`
	ProposedScholarship json.RawMessage `json:"proposedScholarship"`
	Name                string          `json:"name" validate:"required"`
	Country             *string         `json:"country"`
	Intake              *string         `json:"intake"`
	IntakeYear          *int            `json:"intakeYear"`
	Summary             *string         `json:"summary"`
	Priority            string          `json:"priority" validate:"omitempty,oneof=highest high medium low"`
	Rank                *int            `json:"rank" validate:"omitempty,gte=1"`
	Confidence          *float64        `json:"confidence" validate:"omitempty,gte=0,lte=1"`
	ReasoningSummary    *string         `json:"reasoningSummary"`
	SourceIDs           []uuid.UUID     `json:"sourceIds"`
}

type AgentInformationRequestRequest struct {
	ApplicationID      *uuid.UUID      `json:"applicationId"`
	ResearchTaskID     *uuid.UUID      `json:"researchTaskId"`
	QuestionnaireID    *uuid.UUID      `json:"questionnaireId"`
	QuestionID         *uuid.UUID      `json:"questionId"`
	ApplicationFieldID *uuid.UUID      `json:"applicationFieldId"`
	RequestType        string          `json:"requestType" validate:"required"`
	Title              string          `json:"title" validate:"required"`
	Prompt             string          `json:"prompt" validate:"required"`
	Context            *string         `json:"context"`
	Priority           *string         `json:"priority"`
	ResponseType       string          `json:"responseType" validate:"required"`
	Options            json.RawMessage `json:"options"`
}

type ResearchRunDTO struct {
	ID             uuid.UUID  `json:"id"`
	ApplicationID  *uuid.UUID `json:"applicationId,omitempty"`
	ResearchTaskID *uuid.UUID `json:"researchTaskId,omitempty"`
	Status         string     `json:"status"`
	Trigger        string     `json:"trigger"`
	ResearchType   string     `json:"researchType"`
	ModelProvider  *string    `json:"modelProvider,omitempty"`
	ModelName      *string    `json:"modelName,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	ErrorMessage   *string    `json:"errorMessage,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type ResearchSourceDTO struct {
	ID            uuid.UUID  `json:"id"`
	ResearchRunID uuid.UUID  `json:"researchRunId"`
	ApplicationID *uuid.UUID `json:"applicationId,omitempty"`
	URL           string     `json:"url"`
	Title         *string    `json:"title,omitempty"`
	Publisher     *string    `json:"publisher,omitempty"`
	SourceType    string     `json:"sourceType"`
	IsOfficial    bool       `json:"isOfficial"`
	RetrievedAt   time.Time  `json:"retrievedAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type ResearchFindingDTO struct {
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
	CreatedAt          time.Time       `json:"createdAt"`
}

type ProposedInstitution struct {
	Name            string  `json:"name" validate:"required"`
	ShortName       *string `json:"shortName"`
	InstitutionType *string `json:"institutionType"`
	Country         string  `json:"country" validate:"required"`
	City            *string `json:"city"`
	WebsiteURL      *string `json:"websiteUrl"`
}

type ProposedProgramme struct {
	Name           string  `json:"name" validate:"required"`
	DegreeLevel    *string `json:"degreeLevel"`
	FieldOfStudy   *string `json:"fieldOfStudy"`
	Faculty        *string `json:"faculty"`
	Department     *string `json:"department"`
	DurationMonths *int    `json:"durationMonths"`
	Mode           *string `json:"mode"`
	Language       *string `json:"language"`
	ProgrammeURL   *string `json:"programmeUrl"`
	Description    *string `json:"description"`
}

type ProposedScholarship struct {
	Name            string  `json:"name" validate:"required"`
	ProviderName    *string `json:"providerName"`
	Description     *string `json:"description"`
	Country         *string `json:"country"`
	DegreeLevel     *string `json:"degreeLevel"`
	ScholarshipType *string `json:"scholarshipType"`
	OfficialURL     *string `json:"officialUrl"`
	IsRecurring     bool    `json:"isRecurring"`
}

type ApplicationProposalCandidate struct {
	StableKey           string               `json:"stableKey"`
	InstitutionID       *uuid.UUID           `json:"institutionId"`
	ProgrammeID         *uuid.UUID           `json:"programmeId"`
	ScholarshipID       *uuid.UUID           `json:"scholarshipId"`
	ProposedInstitution *ProposedInstitution `json:"proposedInstitution"`
	ProposedProgramme   *ProposedProgramme   `json:"proposedProgramme"`
	ProposedScholarship *ProposedScholarship `json:"proposedScholarship"`
	Name                string               `json:"name"`
	Country             *string              `json:"country"`
	Intake              *string              `json:"intake"`
	IntakeYear          *int                 `json:"intakeYear"`
	Summary             *string              `json:"summary"`
	Priority            string               `json:"priority"`
	Rank                *int                 `json:"rank"`
	Confidence          *float64             `json:"confidence"`
	ReasoningSummary    *string              `json:"reasoningSummary"`
	SourceKeys          []string             `json:"sourceKeys"`
}

type ApplicationProposalResponse struct {
	ID                  uuid.UUID                `json:"id"`
	UserID              uuid.UUID                `json:"userId"`
	ResearchTaskID      uuid.UUID                `json:"researchTaskId"`
	ResearchRunID       *uuid.UUID               `json:"researchRunId,omitempty"`
	InstitutionID       *uuid.UUID               `json:"institutionId,omitempty"`
	ProgrammeID         *uuid.UUID               `json:"programmeId,omitempty"`
	ScholarshipID       *uuid.UUID               `json:"scholarshipId,omitempty"`
	ParentProfileID     *uuid.UUID               `json:"parentProfileId,omitempty"`
	ApplicationID       *uuid.UUID               `json:"applicationId,omitempty"`
	ProposedInstitution json.RawMessage          `json:"proposedInstitution,omitempty"`
	ProposedProgramme   json.RawMessage          `json:"proposedProgramme,omitempty"`
	ProposedScholarship json.RawMessage          `json:"proposedScholarship,omitempty"`
	Name                string                   `json:"name"`
	Country             *string                  `json:"country,omitempty"`
	Intake              *string                  `json:"intake,omitempty"`
	IntakeYear          *int                     `json:"intakeYear,omitempty"`
	Summary             *string                  `json:"summary,omitempty"`
	Status              string                   `json:"status"`
	Priority            string                   `json:"priority"`
	Rank                *int                     `json:"rank,omitempty"`
	Confidence          *float64                 `json:"confidence,omitempty"`
	ReasoningSummary    *string                  `json:"reasoningSummary,omitempty"`
	Sources             []ProposalSourceResponse `json:"sources"`
	CreatedAt           time.Time                `json:"createdAt"`
	UpdatedAt           time.Time                `json:"updatedAt"`
	ReviewedAt          *time.Time               `json:"reviewedAt,omitempty"`
}

type ProposalSourceResponse struct {
	ID               uuid.UUID `json:"id"`
	ResearchSourceID uuid.UUID `json:"researchSourceId"`
	SourceRole       *string   `json:"sourceRole,omitempty"`
	URL              string    `json:"url"`
	Title            *string   `json:"title,omitempty"`
	IsOfficial       bool      `json:"isOfficial"`
}

type ApproveProposalRequest struct {
	ParentProfileID uuid.UUID `json:"parentProfileId" validate:"required"`
}

type ApplicationFieldRequest struct {
	Key            string          `json:"key" validate:"required"`
	Label          string          `json:"label" validate:"required"`
	Value          json.RawMessage `json:"value"`
	ValueType      string          `json:"valueType" validate:"required"`
	Status         *string         `json:"status"`
	SourceType     *string         `json:"sourceType"`
	SourceEntityID *uuid.UUID      `json:"sourceEntityId"`
	Notes          *string         `json:"notes"`
}

type ApplicationFieldResponse struct {
	ID             uuid.UUID       `json:"id"`
	ApplicationID  uuid.UUID       `json:"applicationId"`
	Key            string          `json:"key"`
	Label          string          `json:"label"`
	Value          json.RawMessage `json:"value,omitempty"`
	ValueType      string          `json:"valueType"`
	Status         string          `json:"status"`
	SourceType     *string         `json:"sourceType,omitempty"`
	SourceEntityID *uuid.UUID      `json:"sourceEntityId,omitempty"`
	PrefillRunID   *uuid.UUID      `json:"prefillRunId,omitempty"`
	Notes          *string         `json:"notes,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type QuestionnaireRequest struct {
	Title             string  `json:"title" validate:"required"`
	Description       *string `json:"description"`
	QuestionnaireType *string `json:"questionnaireType"`
	Status            *string `json:"status"`
	SourceURL         *string `json:"sourceUrl" validate:"omitempty,url"`
}

type QuestionnaireResponse struct {
	ID                uuid.UUID          `json:"id"`
	ApplicationID     uuid.UUID          `json:"applicationId"`
	Title             string             `json:"title"`
	Description       *string            `json:"description,omitempty"`
	QuestionnaireType *string            `json:"questionnaireType,omitempty"`
	Status            string             `json:"status"`
	SourceURL         *string            `json:"sourceUrl,omitempty"`
	Questions         []QuestionResponse `json:"questions,omitempty"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`
}

type QuestionRequest struct {
	Key            *string         `json:"key"`
	Prompt         string          `json:"prompt" validate:"required"`
	HelpText       *string         `json:"helpText"`
	QuestionType   string          `json:"questionType" validate:"required"`
	IsRequired     bool            `json:"isRequired"`
	WordLimit      *int            `json:"wordLimit" validate:"omitempty,gte=1"`
	CharacterLimit *int            `json:"characterLimit" validate:"omitempty,gte=1"`
	SortOrder      *int            `json:"sortOrder"`
	Options        json.RawMessage `json:"options"`
	Status         *string         `json:"status"`
}

type QuestionResponse struct {
	ID              uuid.UUID       `json:"id"`
	QuestionnaireID uuid.UUID       `json:"questionnaireId"`
	Key             *string         `json:"key,omitempty"`
	Prompt          string          `json:"prompt"`
	HelpText        *string         `json:"helpText,omitempty"`
	QuestionType    string          `json:"questionType"`
	IsRequired      bool            `json:"isRequired"`
	WordLimit       *int            `json:"wordLimit,omitempty"`
	CharacterLimit  *int            `json:"characterLimit,omitempty"`
	SortOrder       *int            `json:"sortOrder,omitempty"`
	Options         json.RawMessage `json:"options,omitempty"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type AnswerRequest struct {
	Value            json.RawMessage `json:"value"`
	DraftText        *string         `json:"draftText"`
	Status           *string         `json:"status"`
	AnswerSource     *string         `json:"answerSource"`
	SourceEntityType *string         `json:"sourceEntityType"`
	SourceEntityID   *uuid.UUID      `json:"sourceEntityId"`
	Confidence       *float64        `json:"confidence" validate:"omitempty,gte=0,lte=1"`
	CreatedBy        *string         `json:"createdBy"`
}

type AnswerResponse struct {
	ID               uuid.UUID       `json:"id"`
	QuestionID       uuid.UUID       `json:"questionId"`
	Value            json.RawMessage `json:"value,omitempty"`
	DraftText        *string         `json:"draftText,omitempty"`
	Status           string          `json:"status"`
	AnswerSource     *string         `json:"answerSource,omitempty"`
	SourceEntityType *string         `json:"sourceEntityType,omitempty"`
	SourceEntityID   *uuid.UUID      `json:"sourceEntityId,omitempty"`
	Confidence       *float64        `json:"confidence,omitempty"`
	CreatedBy        *string         `json:"createdBy,omitempty"`
	PrefillRunID     *uuid.UUID      `json:"prefillRunId,omitempty"`
	AgentProvider    *string         `json:"agentProvider,omitempty"`
	AgentModel       *string         `json:"agentModel,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type InformationRequestFilters struct {
	UserID        *uuid.UUID
	Status        *string
	ApplicationID *uuid.UUID
	RequestType   *string
}

type CreateInformationRequest struct {
	UserID             uuid.UUID       `json:"userId" validate:"required"`
	ApplicationID      *uuid.UUID      `json:"applicationId"`
	ResearchTaskID     *uuid.UUID      `json:"researchTaskId"`
	QuestionnaireID    *uuid.UUID      `json:"questionnaireId"`
	QuestionID         *uuid.UUID      `json:"questionId"`
	ApplicationFieldID *uuid.UUID      `json:"applicationFieldId"`
	RequestType        string          `json:"requestType" validate:"required"`
	Title              string          `json:"title" validate:"required"`
	Prompt             string          `json:"prompt" validate:"required"`
	Context            *string         `json:"context"`
	Priority           *string         `json:"priority"`
	ResponseType       string          `json:"responseType" validate:"required"`
	Options            json.RawMessage `json:"options"`
	CreatedBy          string          `json:"createdBy" validate:"required"`
}

type RespondInformationRequest struct {
	ResponseValue json.RawMessage `json:"responseValue"`
	ResponseText  *string         `json:"responseText"`
	SubmittedBy   string          `json:"submittedBy" validate:"required"`
}

type InformationRequestResponseDTO struct {
	ID                 uuid.UUID                `json:"id"`
	UserID             uuid.UUID                `json:"userId"`
	ApplicationID      *uuid.UUID               `json:"applicationId,omitempty"`
	ResearchTaskID     *uuid.UUID               `json:"researchTaskId,omitempty"`
	QuestionnaireID    *uuid.UUID               `json:"questionnaireId,omitempty"`
	QuestionID         *uuid.UUID               `json:"questionId,omitempty"`
	ApplicationFieldID *uuid.UUID               `json:"applicationFieldId,omitempty"`
	RequestType        string                   `json:"requestType"`
	Title              string                   `json:"title"`
	Prompt             string                   `json:"prompt"`
	Context            *string                  `json:"context,omitempty"`
	Status             string                   `json:"status"`
	Priority           *string                  `json:"priority,omitempty"`
	ResponseType       string                   `json:"responseType"`
	Options            json.RawMessage          `json:"options,omitempty"`
	ResponseValue      json.RawMessage          `json:"responseValue,omitempty"`
	ResponseText       *string                  `json:"responseText,omitempty"`
	CreatedBy          string                   `json:"createdBy"`
	ResolutionSource   *string                  `json:"resolutionSource,omitempty"`
	Responses          []InformationResponseDTO `json:"responses"`
	CreatedAt          time.Time                `json:"createdAt"`
	UpdatedAt          time.Time                `json:"updatedAt"`
	CompletedAt        *time.Time               `json:"completedAt,omitempty"`
	ReopenedAt         *time.Time               `json:"reopenedAt,omitempty"`
}

type InformationResponseDTO struct {
	ID            uuid.UUID       `json:"id"`
	ResponseValue json.RawMessage `json:"responseValue,omitempty"`
	ResponseText  *string         `json:"responseText,omitempty"`
	SubmittedBy   string          `json:"submittedBy"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type PrefillRequest struct {
	Trigger    string `json:"trigger" validate:"omitempty,oneof=application_created manual information_updated questionnaire_added regeneration"`
	Regenerate bool   `json:"regenerate"`
}

type PrefillRunResponse struct {
	ID             uuid.UUID               `json:"id"`
	ApplicationID  uuid.UUID               `json:"applicationId"`
	ResearchTaskID *uuid.UUID              `json:"researchTaskId,omitempty"`
	Status         string                  `json:"status"`
	Trigger        string                  `json:"trigger"`
	AgentProvider  *string                 `json:"agentProvider,omitempty"`
	AgentModel     *string                 `json:"agentModel,omitempty"`
	StartedAt      *time.Time              `json:"startedAt,omitempty"`
	CompletedAt    *time.Time              `json:"completedAt,omitempty"`
	ErrorMessage   *string                 `json:"errorMessage,omitempty"`
	Actions        []PrefillActionResponse `json:"actions"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
}

type PrefillActionResponse struct {
	ID         uuid.UUID  `json:"id"`
	ActionType string     `json:"actionType"`
	TargetType string     `json:"targetType"`
	TargetID   *uuid.UUID `json:"targetId,omitempty"`
	Status     string     `json:"status"`
	Summary    *string    `json:"summary,omitempty"`
	SourceType *string    `json:"sourceType,omitempty"`
	SourceID   *uuid.UUID `json:"sourceId,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type PreparationSummary struct {
	FieldsCompleted           int `json:"fieldsCompleted"`
	FieldsTotal               int `json:"fieldsTotal"`
	QuestionnairesCompleted   int `json:"questionnairesCompleted"`
	QuestionnairesTotal       int `json:"questionnairesTotal"`
	PendingInformation        int `json:"pendingInformation"`
	RequiredQuestionsAnswered int `json:"requiredQuestionsAnswered"`
	RequiredQuestionsTotal    int `json:"requiredQuestionsTotal"`
}
