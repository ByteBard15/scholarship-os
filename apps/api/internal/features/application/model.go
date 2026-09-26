package application

import (
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/catalog"
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

type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Status string

const (
	StatusDiscovered          Status = "discovered"
	StatusResearching         Status = "researching"
	StatusResearchComplete    Status = "research_complete"
	StatusEligibilityReview   Status = "eligibility_review"
	StatusPreparing           Status = "preparing"
	StatusDocumentsInProgress Status = "documents_in_progress"
	StatusReadyToSubmit       Status = "ready_to_submit"
	StatusSubmitted           Status = "submitted"
	StatusWaiting             Status = "waiting"
	StatusAccepted            Status = "accepted"
	StatusRejected            Status = "rejected"
	StatusWithdrawn           Status = "withdrawn"
	StatusExpired             Status = "expired"
)

func (s Status) Valid() bool { _, ok := statusTransitions[s]; return ok }

var statusTransitions = map[Status]map[Status]bool{
	StatusDiscovered:          {StatusResearching: true, StatusWithdrawn: true, StatusExpired: true},
	StatusResearching:         {StatusResearchComplete: true, StatusWithdrawn: true, StatusExpired: true},
	StatusResearchComplete:    {StatusEligibilityReview: true, StatusResearching: true, StatusWithdrawn: true, StatusExpired: true},
	StatusEligibilityReview:   {StatusPreparing: true, StatusResearching: true, StatusWithdrawn: true, StatusExpired: true},
	StatusPreparing:           {StatusDocumentsInProgress: true, StatusWithdrawn: true, StatusExpired: true},
	StatusDocumentsInProgress: {StatusReadyToSubmit: true, StatusPreparing: true, StatusWithdrawn: true, StatusExpired: true},
	StatusReadyToSubmit:       {StatusSubmitted: true, StatusDocumentsInProgress: true, StatusWithdrawn: true, StatusExpired: true},
	StatusSubmitted:           {StatusWaiting: true, StatusWithdrawn: true},
	StatusWaiting:             {StatusAccepted: true, StatusRejected: true, StatusWithdrawn: true},
	StatusAccepted:            {}, StatusRejected: {}, StatusWithdrawn: {}, StatusExpired: {},
}

func CanTransition(from, to Status) bool { return from == to || statusTransitions[from][to] }

type ResearchStatus string

const (
	ResearchNotStarted     ResearchStatus = "not_started"
	ResearchQueued         ResearchStatus = "queued"
	ResearchRunning        ResearchStatus = "running"
	ResearchReviewRequired ResearchStatus = "review_required"
	ResearchComplete       ResearchStatus = "complete"
	ResearchFailed         ResearchStatus = "failed"
	ResearchStale          ResearchStatus = "stale"
)

type Application struct {
	Base
	UserID             uuid.UUID  `gorm:"type:uuid;not null;index"`
	ApplicantProfileID uuid.UUID  `gorm:"type:uuid;not null;index"`
	InstitutionID      *uuid.UUID `gorm:"type:uuid;index"`
	ProgrammeID        *uuid.UUID `gorm:"type:uuid;index"`
	ScholarshipID      *uuid.UUID `gorm:"type:uuid;index"`
	Name               string     `gorm:"not null"`
	Intake             *string
	IntakeYear         *int
	Country            *string
	Status             Status `gorm:"type:text;not null"`
	Priority           *int
	Notes              *string        `gorm:"type:text"`
	ResearchStatus     ResearchStatus `gorm:"type:text;not null"`
	SoftDelete
	Institution *catalog.Institution `gorm:"foreignKey:InstitutionID"`
	Programme   *catalog.Programme   `gorm:"foreignKey:ProgrammeID"`
	Scholarship *catalog.Scholarship `gorm:"foreignKey:ScholarshipID"`
}

type ApplicationRequirement struct {
	Base
	ApplicationID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Category         string    `gorm:"not null;index"`
	Title            string    `gorm:"not null"`
	Description      *string   `gorm:"type:text"`
	IsMandatory      bool
	Status           string     `gorm:"not null;index"`
	DueDate          *time.Time `gorm:"type:date"`
	SourceID         *uuid.UUID `gorm:"type:uuid"`
	SourceURL        *string
	EvidenceRequired bool
	Notes            *string
	SortOrder        *int
	SoftDelete
}
type ApplicationDeadline struct {
	Base
	ApplicationID   uuid.UUID `gorm:"type:uuid;not null;index"`
	DeadlineType    string    `gorm:"not null;index"`
	Title           string    `gorm:"not null"`
	DeadlineAt      *time.Time
	Timezone        *string
	DatePrecision   string `gorm:"not null"`
	RawDeadlineText *string
	IsHardDeadline  bool
	SourceID        *uuid.UUID `gorm:"type:uuid"`
	SourceURL       *string
	VerifiedAt      *time.Time
	Notes           *string
	SoftDelete
}
type ApplicationFunding struct {
	Base
	ApplicationID         uuid.UUID `gorm:"type:uuid;not null;index"`
	FundingType           string    `gorm:"not null"`
	Currency              *string
	Amount                *float64
	AmountPeriod          *string
	TuitionCoverage       *string
	StipendAmount         *float64
	StipendPeriod         *string
	TravelCoverage        *string
	InsuranceCoverage     *string
	AccommodationCoverage *string
	OtherBenefits         *string    `gorm:"type:text"`
	Conditions            *string    `gorm:"type:text"`
	SourceID              *uuid.UUID `gorm:"type:uuid"`
	SourceURL             *string
	SoftDelete
}
type ApplicationContact struct {
	Base
	ApplicationID uuid.UUID `gorm:"type:uuid;not null;index"`
	Name          *string
	Role          *string
	Email         *string
	Phone         *string
	Organization  *string
	ContactType   *string
	URL           *string
	Notes         *string
	SourceID      *uuid.UUID `gorm:"type:uuid"`
	SoftDelete
}
type ApplicationSupervisor struct {
	Base
	ApplicationID uuid.UUID `gorm:"type:uuid;not null;index"`
	Name          string    `gorm:"not null"`
	Title         *string
	Department    *string
	Institution   *string
	Email         *string
	ProfileURL    *string
	ResearchAreas *string `gorm:"type:text"`
	ContactStatus *string
	Notes         *string
	SourceID      *uuid.UUID `gorm:"type:uuid"`
	SoftDelete
}
type ApplicationURL struct {
	Base
	ApplicationID uuid.UUID `gorm:"type:uuid;not null;index"`
	URLType       string    `gorm:"not null"`
	Label         *string
	URL           string `gorm:"not null"`
	IsOfficial    bool
	SoftDelete
}
type ApplicationTask struct {
	Base
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentTaskID  *uuid.UUID `gorm:"type:uuid;index"`
	Title         string     `gorm:"not null"`
	Description   *string
	Status        string `gorm:"not null;index"`
	Priority      *int
	DueAt         *time.Time
	CompletedAt   *time.Time
	SortOrder     *int
	TaskType      *string
	Source        *string
	SoftDelete
}
type ExternalTaskReference struct {
	Base
	ApplicationTaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Provider          string    `gorm:"not null"`
	ExternalListID    *string
	ExternalTaskID    string `gorm:"not null"`
	ExternalURL       *string
	LastSyncedAt      *time.Time
	SyncStatus        *string
}
type RequirementEvidence struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	RequirementID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ProfileID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	EntityType    string     `gorm:"not null"`
	EntityID      uuid.UUID  `gorm:"type:uuid;not null"`
	EvidenceID    *uuid.UUID `gorm:"type:uuid"`
	Status        string     `gorm:"not null"`
	Notes         *string
	CreatedAt     time.Time
}

func (v *RequirementEvidence) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type ResearchRun struct {
	Base
	ApplicationID  *uuid.UUID `gorm:"type:uuid;index"`
	ResearchTaskID *uuid.UUID `gorm:"type:uuid;index"`
	Status         string     `gorm:"not null;index"`
	Trigger        string     `gorm:"not null"`
	ResearchType   string     `gorm:"not null"`
	ModelProvider  *string
	ModelName      *string
	PromptVersion  *string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	ErrorMessage   *string
}
type ResearchSource struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ResearchRunID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ApplicationID *uuid.UUID `gorm:"type:uuid;index"`
	URL           string     `gorm:"not null"`
	Title         *string
	Publisher     *string
	SourceType    string `gorm:"not null"`
	IsOfficial    bool
	RetrievedAt   time.Time
	PublishedAt   *time.Time
	ContentHash   *string
	Notes         *string
	CreatedAt     time.Time
}

func (v *ResearchSource) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type ResearchFinding struct {
	ID                 uuid.UUID    `gorm:"type:uuid;primaryKey"`
	ResearchRunID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	ApplicationID      *uuid.UUID   `gorm:"type:uuid;index"`
	SourceID           *uuid.UUID   `gorm:"type:uuid"`
	Category           string       `gorm:"not null;index"`
	Field              string       `gorm:"not null"`
	Value              profile.JSON `gorm:"type:jsonb;not null"`
	Confidence         *float64
	VerificationStatus string  `gorm:"not null"`
	ReviewStatus       string  `gorm:"not null;index"`
	RawText            *string `gorm:"type:text"`
	Notes              *string
	CreatedAt          time.Time
}

func (v *ResearchFinding) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

func (ApplicationRequirement) TableName() string { return "application_requirements" }
func (ApplicationDeadline) TableName() string    { return "application_deadlines" }
func (ApplicationFunding) TableName() string     { return "application_funding" }
func (ApplicationContact) TableName() string     { return "application_contacts" }
func (ApplicationSupervisor) TableName() string  { return "application_supervisors" }
func (ApplicationURL) TableName() string         { return "application_urls" }
func (ApplicationTask) TableName() string        { return "application_tasks" }
func (ExternalTaskReference) TableName() string  { return "external_task_references" }
func (RequirementEvidence) TableName() string    { return "requirement_evidence" }
func (ResearchRun) TableName() string            { return "research_runs" }
func (ResearchSource) TableName() string         { return "research_sources" }
func (ResearchFinding) TableName() string        { return "research_findings" }
