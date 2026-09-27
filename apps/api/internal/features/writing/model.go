package writing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TypePersonalStatement    = "personal_statement"
	TypeMotivationLetter     = "motivation_letter"
	TypeRecommendationLetter = "recommendation_letter"
	TypeStatementOfPurpose   = "statement_of_purpose"
	TypeEssay                = "essay"
	TypeCoverLetter          = "cover_letter"
	TypeOther                = "other"
	OriginUser               = "user"
	OriginAgent              = "agent"
	OriginImport             = "import"
	StatusSource             = "source"
	StatusDraft              = "draft"
	StatusSuggested          = "suggested"
	StatusApproved           = "approved"
	StatusArchived           = "archived"
)

type WritingSample struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	ApplicationID *uuid.UUID `gorm:"type:uuid;index"`
	Title         string     `gorm:"not null"`
	DocumentType  string     `gorm:"not null;index"`
	Content       string     `gorm:"type:text;not null"`
	Origin        string     `gorm:"not null"`
	Status        string     `gorm:"not null;index"`
	Description   *string    `gorm:"type:text"`
	CreatedBy     *string
	Tags          []WritingTag `gorm:"many2many:writing_sample_tags"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (WritingSample) TableName() string { return "writing_samples" }

type WritingTag struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Name           string    `gorm:"not null"`
	NormalizedName string    `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (WritingTag) TableName() string { return "writing_tags" }

type WritingSampleRelation struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	SourceSampleID  uuid.UUID `gorm:"type:uuid;not null;index"`
	DerivedSampleID uuid.UUID `gorm:"type:uuid;not null;index"`
	RelationType    string    `gorm:"not null"`
	CreatedAt       time.Time
}

func (WritingSampleRelation) TableName() string { return "writing_sample_relations" }

type WritingSampleTag struct {
	WritingSampleID uuid.UUID `gorm:"type:uuid;primaryKey"`
	WritingTagID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt       time.Time
}

func (WritingSampleTag) TableName() string { return "writing_sample_tags" }

func (sample *WritingSample) BeforeCreate(_ *gorm.DB) error {
	if sample.ID == uuid.Nil {
		sample.ID = uuid.New()
	}
	return nil
}
