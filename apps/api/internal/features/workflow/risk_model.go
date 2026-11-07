package workflow

import (
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"time"
)

type ProposalRiskAssessment struct {
	ID                uuid.UUID    `gorm:"type:uuid;primaryKey"`
	ProposalID        uuid.UUID    `gorm:"type:uuid;not null"`
	ProfileSnapshotID uuid.UUID    `gorm:"type:uuid;not null"`
	ProposalSnapshot  profile.JSON `gorm:"type:jsonb;not null"`
	RiskFactor        *float64
	Methodology       string
	Notes             string
	RecordedBy        *uuid.UUID `gorm:"type:uuid"`
	RecordedByType    string
	ComputedAt        time.Time
	CreatedAt         time.Time
	Variables         []ProposalRiskVariable `gorm:"foreignKey:AssessmentID"`
}

func (ProposalRiskAssessment) TableName() string { return "proposal_risk_assessments" }

type ProposalRiskVariable struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	AssessmentID  uuid.UUID `gorm:"type:uuid;not null"`
	VariableName  string
	Value         *float64
	Rationale     string
	SourceURL     *string
	ObservedAt    time.Time
	TimeVariable  bool
	TimePredicate *string
	ValidUntil    *time.Time
	Blocking      bool
}

func (ProposalRiskVariable) TableName() string { return "proposal_risk_variables" }
