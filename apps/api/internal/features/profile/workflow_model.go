package profile

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch raw := value.(type) {
	case []byte:
		*j = append((*j)[:0], raw...)
	case string:
		*j = append((*j)[:0], raw...)
	default:
		return errors.New("unsupported JSON database value")
	}
	return nil
}

type ProfileType string

const (
	ProfileTypeMaster      ProfileType = "master"
	ProfileTypeDomain      ProfileType = "domain"
	ProfileTypeApplication ProfileType = "application"
)

func (p ProfileType) Valid() bool {
	return p == ProfileTypeMaster || p == ProfileTypeDomain || p == ProfileTypeApplication
}

type OverrideType string

const (
	OverrideReplace OverrideType = "replace"
	OverrideHide    OverrideType = "hide"
	OverrideAppend  OverrideType = "append"
)

type ProfileOverride struct {
	Base
	ProfileID    uuid.UUID    `gorm:"type:uuid;not null;index"`
	EntityType   string       `gorm:"not null;index:idx_profile_override_entity"`
	EntityID     *uuid.UUID   `gorm:"type:uuid;index:idx_profile_override_entity"`
	FieldName    string       `gorm:"not null"`
	OverrideType OverrideType `gorm:"type:text;not null"`
	Value        JSON         `gorm:"type:jsonb"`
	Reason       *string      `gorm:"type:text"`
	Source       *string
	SoftDelete
}

func (ProfileOverride) TableName() string { return "profile_overrides" }

type ProfileSnapshot struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProfileID uuid.UUID `gorm:"type:uuid;not null;index"`
	Version   int       `gorm:"not null"`
	Snapshot  JSON      `gorm:"type:jsonb;not null"`
	CreatedAt time.Time
	CreatedBy *string
	Reason    *string
}

func (ProfileSnapshot) TableName() string { return "profile_snapshots" }
func (s *ProfileSnapshot) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type ProfileDocument struct {
	Base
	ProfileID        uuid.UUID `gorm:"type:uuid;not null;index"`
	DocumentType     string    `gorm:"not null;index"`
	OriginalFilename string    `gorm:"not null"`
	StorageProvider  string    `gorm:"not null"`
	StorageKey       string    `gorm:"not null;uniqueIndex"`
	MimeType         string    `gorm:"not null"`
	FileSize         int64     `gorm:"not null"`
	SHA256           *string
	Status           string `gorm:"not null"`
}

func (ProfileDocument) TableName() string { return "profile_documents" }

type ProfileImport struct {
	Base
	ProfileID          uuid.UUID `gorm:"type:uuid;not null;index"`
	DocumentID         uuid.UUID `gorm:"type:uuid;not null;index"`
	ImportType         string    `gorm:"not null"`
	Status             string    `gorm:"not null;index"`
	RawText            *string   `gorm:"type:text"`
	ExtractedData      JSON      `gorm:"type:jsonb"`
	ExtractionProvider *string
	ExtractionModel    *string
	ErrorMessage       *string `gorm:"type:text"`
	CompletedAt        *time.Time
}

func (ProfileImport) TableName() string { return "profile_imports" }

type ProfileImportCandidate struct {
	Base
	ImportID        uuid.UUID `gorm:"type:uuid;not null;index"`
	SectionType     string    `gorm:"not null;index"`
	CandidateData   JSON      `gorm:"type:jsonb;not null"`
	SourceText      *string   `gorm:"type:text"`
	Confidence      *float64
	Status          string     `gorm:"not null;index"`
	MatchedEntityID *uuid.UUID `gorm:"type:uuid"`
}

func (ProfileImportCandidate) TableName() string { return "profile_import_candidates" }

type ProfileEvidence struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProfileID    uuid.UUID `gorm:"type:uuid;not null;index"`
	EntityType   string    `gorm:"not null;index:idx_profile_evidence_entity"`
	EntityID     uuid.UUID `gorm:"type:uuid;not null;index:idx_profile_evidence_entity"`
	FieldName    *string
	EvidenceType string     `gorm:"not null"`
	DocumentID   *uuid.UUID `gorm:"type:uuid"`
	ImportID     *uuid.UUID `gorm:"type:uuid"`
	SourceURL    *string
	SourceText   *string `gorm:"type:text"`
	Confidence   *float64
	CreatedAt    time.Time
}

func (ProfileEvidence) TableName() string { return "profile_evidence" }
func (e *ProfileEvidence) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID     *uuid.UUID `gorm:"type:uuid;index"`
	ProfileID  *uuid.UUID `gorm:"type:uuid;index"`
	Action     string     `gorm:"not null"`
	EntityType *string
	EntityID   *uuid.UUID `gorm:"type:uuid"`
	Metadata   JSON       `gorm:"type:jsonb"`
	CreatedAt  time.Time
}

func (AuditLog) TableName() string { return "audit_logs" }
func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
