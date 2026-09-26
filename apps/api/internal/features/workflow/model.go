package workflow

import (
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type ResearchTask struct {
	Base
	UserID              uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentTaskID        *uuid.UUID `gorm:"type:uuid;index"`
	TargetApplicationID *uuid.UUID `gorm:"type:uuid;index"`
	Title               string     `gorm:"not null"`
	Description         *string    `gorm:"type:text"`
	Instructions        *string    `gorm:"type:text"`
	TaskType            string     `gorm:"not null;index"`
	Status              string     `gorm:"not null;index"`
	Priority            *string
	DueAt               *time.Time
	ResearchConfig      profile.JSON `gorm:"type:jsonb"`
	StartedAt           *time.Time
	CompletedAt         *time.Time
	FailedAt            *time.Time
	FailureReason       *string `gorm:"type:text"`
	DeletedAt           gorm.DeletedAt
}

type ResearchTaskLink struct {
	Base
	ResearchTaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Label          *string
	URL            string `gorm:"not null"`
	LinkType       *string
}

// ResearchContext is reusable user-owned background information supplied to
// research agents. It is deliberately separate from InformationRequest, which
// represents missing information in an application workflow.
type ResearchContext struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Question  string    `gorm:"type:text;not null"`
	Answer    string    `gorm:"type:text;not null"`
	DeletedAt gorm.DeletedAt
}

type ResearchTaskContext struct {
	ResearchTaskID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	ResearchContextID uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt         time.Time
}

type ResearchTaskOutput struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	ResearchTaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	OutputType     string    `gorm:"not null"`
	EntityID       uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt      time.Time
}

func (v *ResearchTaskOutput) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type ApplicationProposal struct {
	Base
	UserID              uuid.UUID    `gorm:"type:uuid;not null;index"`
	ResearchTaskID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	ResearchRunID       *uuid.UUID   `gorm:"type:uuid"`
	InstitutionID       *uuid.UUID   `gorm:"type:uuid"`
	ProgrammeID         *uuid.UUID   `gorm:"type:uuid"`
	ScholarshipID       *uuid.UUID   `gorm:"type:uuid"`
	ProposedInstitution profile.JSON `gorm:"type:jsonb"`
	ProposedProgramme   profile.JSON `gorm:"type:jsonb"`
	ProposedScholarship profile.JSON `gorm:"type:jsonb"`
	Name                string       `gorm:"not null"`
	Country             *string
	Intake              *string
	IntakeYear          *int
	Summary             *string `gorm:"type:text"`
	Status              string  `gorm:"not null;index"`
	Confidence          *float64
	ReasoningSummary    *string `gorm:"type:text"`
	ReviewedAt          *time.Time
}

type ApplicationProposalSource struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey"`
	ApplicationProposalID uuid.UUID `gorm:"type:uuid;not null;index"`
	ResearchSourceID      uuid.UUID `gorm:"type:uuid;not null"`
	SourceRole            *string
	CreatedAt             time.Time
}

func (v *ApplicationProposalSource) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type ApplicationField struct {
	Base
	ApplicationID  uuid.UUID    `gorm:"type:uuid;not null;index"`
	Key            string       `gorm:"not null"`
	Label          string       `gorm:"not null"`
	Value          profile.JSON `gorm:"type:jsonb"`
	ValueType      string       `gorm:"not null"`
	Status         string       `gorm:"not null;index"`
	SourceType     *string
	SourceEntityID *uuid.UUID `gorm:"type:uuid"`
	PrefillRunID   *uuid.UUID `gorm:"type:uuid"`
	Notes          *string
}

type ApplicationQuestionnaire struct {
	Base
	ApplicationID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Title             string    `gorm:"not null"`
	Description       *string   `gorm:"type:text"`
	QuestionnaireType *string
	Status            string `gorm:"not null;index"`
	SourceURL         *string
	DeletedAt         gorm.DeletedAt
}

type ApplicationQuestion struct {
	Base
	QuestionnaireID uuid.UUID `gorm:"type:uuid;not null;index"`
	Key             *string
	Prompt          string  `gorm:"type:text;not null"`
	HelpText        *string `gorm:"type:text"`
	QuestionType    string  `gorm:"not null"`
	IsRequired      bool
	WordLimit       *int
	CharacterLimit  *int
	SortOrder       *int
	Options         profile.JSON `gorm:"type:jsonb"`
	Status          string       `gorm:"not null;index"`
	DeletedAt       gorm.DeletedAt
}

type ApplicationAnswer struct {
	Base
	QuestionID       uuid.UUID    `gorm:"type:uuid;not null;index"`
	Value            profile.JSON `gorm:"type:jsonb"`
	DraftText        *string      `gorm:"type:text"`
	Status           string       `gorm:"not null;index"`
	AnswerSource     *string
	SourceEntityType *string
	SourceEntityID   *uuid.UUID `gorm:"type:uuid"`
	Confidence       *float64
	CreatedBy        *string
	PrefillRunID     *uuid.UUID `gorm:"type:uuid"`
	AgentProvider    *string
	AgentModel       *string
}

type InformationRequest struct {
	Base
	UserID             uuid.UUID  `gorm:"type:uuid;not null;index"`
	ApplicationID      *uuid.UUID `gorm:"type:uuid;index"`
	ResearchTaskID     *uuid.UUID `gorm:"type:uuid"`
	QuestionnaireID    *uuid.UUID `gorm:"type:uuid"`
	QuestionID         *uuid.UUID `gorm:"type:uuid;index"`
	ApplicationFieldID *uuid.UUID `gorm:"type:uuid;index"`
	RequestType        string     `gorm:"not null"`
	Title              string     `gorm:"not null"`
	Prompt             string     `gorm:"type:text;not null"`
	Context            *string    `gorm:"type:text"`
	Status             string     `gorm:"not null;index"`
	Priority           *string
	ResponseType       string       `gorm:"not null"`
	Options            profile.JSON `gorm:"type:jsonb"`
	ResponseValue      profile.JSON `gorm:"type:jsonb"`
	ResponseText       *string      `gorm:"type:text"`
	CreatedBy          string       `gorm:"not null"`
	ResolutionSource   *string
	CompletedAt        *time.Time
	ReopenedAt         *time.Time
}

type InformationRequestResponse struct {
	ID                   uuid.UUID    `gorm:"type:uuid;primaryKey"`
	InformationRequestID uuid.UUID    `gorm:"type:uuid;not null;index"`
	ResponseValue        profile.JSON `gorm:"type:jsonb"`
	ResponseText         *string      `gorm:"type:text"`
	SubmittedBy          string       `gorm:"not null"`
	CreatedAt            time.Time
}

func (v *InformationRequestResponse) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type ApplicationPrefillRun struct {
	Base
	ApplicationID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	ResearchTaskID *uuid.UUID `gorm:"type:uuid"`
	Status         string     `gorm:"not null;index"`
	Trigger        string     `gorm:"not null"`
	AgentProvider  *string
	AgentModel     *string
	PromptVersion  *string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	ErrorMessage   *string `gorm:"type:text"`
}

type PrefillAction struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PrefillRunID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ActionType   string     `gorm:"not null"`
	TargetType   string     `gorm:"not null"`
	TargetID     *uuid.UUID `gorm:"type:uuid"`
	Status       string     `gorm:"not null"`
	Summary      *string
	SourceType   *string
	SourceID     *uuid.UUID `gorm:"type:uuid"`
	CreatedAt    time.Time
}

func (v *PrefillAction) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type AgentActivity struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ResearchTaskID *uuid.UUID `gorm:"type:uuid;index"`
	ApplicationID  *uuid.UUID `gorm:"type:uuid;index"`
	PrefillRunID   *uuid.UUID `gorm:"type:uuid"`
	ActivityType   string     `gorm:"not null"`
	Summary        string     `gorm:"not null"`
	CreatedAt      time.Time
}

func (v *AgentActivity) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

func (ResearchTask) TableName() string               { return "research_tasks" }
func (ResearchTaskLink) TableName() string           { return "research_task_links" }
func (ResearchContext) TableName() string            { return "research_contexts" }
func (ResearchTaskContext) TableName() string        { return "research_task_contexts" }
func (ResearchTaskOutput) TableName() string         { return "research_task_outputs" }
func (ApplicationProposal) TableName() string        { return "application_proposals" }
func (ApplicationProposalSource) TableName() string  { return "application_proposal_sources" }
func (ApplicationField) TableName() string           { return "application_fields" }
func (ApplicationQuestionnaire) TableName() string   { return "application_questionnaires" }
func (ApplicationQuestion) TableName() string        { return "application_questions" }
func (ApplicationAnswer) TableName() string          { return "application_answers" }
func (InformationRequest) TableName() string         { return "information_requests" }
func (InformationRequestResponse) TableName() string { return "information_request_responses" }
func (ApplicationPrefillRun) TableName() string      { return "application_prefill_runs" }
func (PrefillAction) TableName() string              { return "prefill_actions" }
func (AgentActivity) TableName() string              { return "agent_activities" }
