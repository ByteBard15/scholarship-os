package profile

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LineageItemResponse struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	ProfileType ProfileType `json:"type"`
}
type EffectiveProfileResponse struct {
	FullProfileResponse
	Lineage []LineageItemResponse `json:"lineage"`
}

type CreateOverrideRequest struct {
	EntityType   string          `json:"entityType" validate:"required"`
	EntityID     *uuid.UUID      `json:"entityId"`
	FieldName    string          `json:"fieldName" validate:"required"`
	OverrideType OverrideType    `json:"overrideType" validate:"required"`
	Value        json.RawMessage `json:"value"`
	Reason       *string         `json:"reason"`
	Source       *string         `json:"source"`
}
type UpdateOverrideRequest struct {
	Value  json.RawMessage `json:"value"`
	Reason *string         `json:"reason"`
	Source *string         `json:"source"`
}
type OverrideResponse struct {
	ID           uuid.UUID       `json:"id"`
	ProfileID    uuid.UUID       `json:"profileId"`
	EntityType   string          `json:"entityType"`
	EntityID     *uuid.UUID      `json:"entityId,omitempty"`
	FieldName    string          `json:"fieldName"`
	OverrideType OverrideType    `json:"overrideType"`
	Value        json.RawMessage `json:"value,omitempty"`
	Reason       *string         `json:"reason,omitempty"`
	Source       *string         `json:"source,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type CreateSnapshotRequest struct {
	CreatedBy *string `json:"createdBy"`
	Reason    *string `json:"reason"`
}
type SnapshotResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProfileID uuid.UUID       `json:"profileId"`
	Version   int             `json:"version"`
	Snapshot  json.RawMessage `json:"snapshot,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	CreatedBy *string         `json:"createdBy,omitempty"`
	Reason    *string         `json:"reason,omitempty"`
}

type DocumentResponse struct {
	ID               uuid.UUID `json:"id"`
	ProfileID        uuid.UUID `json:"profileId"`
	DocumentType     string    `json:"documentType"`
	OriginalFilename string    `json:"originalFilename"`
	StorageProvider  string    `json:"storageProvider"`
	MimeType         string    `json:"mimeType"`
	FileSize         int64     `json:"fileSize"`
	SHA256           *string   `json:"sha256,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type CreateImportRequest struct {
	DocumentID uuid.UUID `json:"documentId" validate:"required"`
}
type ImportResponse struct {
	ID                 uuid.UUID       `json:"id"`
	ProfileID          uuid.UUID       `json:"profileId"`
	DocumentID         uuid.UUID       `json:"documentId"`
	ImportType         string          `json:"importType"`
	Status             string          `json:"status"`
	ExtractedData      json.RawMessage `json:"extractedData,omitempty"`
	ExtractionProvider *string         `json:"extractionProvider,omitempty"`
	ExtractionModel    *string         `json:"extractionModel,omitempty"`
	ErrorMessage       *string         `json:"errorMessage,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	CompletedAt        *time.Time      `json:"completedAt,omitempty"`
}
type CandidateActionRequest struct {
	Action string `json:"action" validate:"required,oneof=accept reject merge"`
}
type FieldChangeResponse struct {
	Field     string `json:"field"`
	Existing  any    `json:"existing"`
	Candidate any    `json:"candidate"`
}
type ImportCandidateResponse struct {
	ID              uuid.UUID             `json:"id"`
	ImportID        uuid.UUID             `json:"importId"`
	SectionType     string                `json:"sectionType"`
	CandidateData   json.RawMessage       `json:"candidateData"`
	SourceText      *string               `json:"sourceText,omitempty"`
	Confidence      *float64              `json:"confidence,omitempty"`
	Status          string                `json:"status"`
	MatchedEntityID *uuid.UUID            `json:"matchedEntityId,omitempty"`
	Existing        any                   `json:"existing,omitempty"`
	Changes         []FieldChangeResponse `json:"changes,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

type CompletenessSectionResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
type CompletenessResponse struct {
	Score    int                           `json:"score"`
	Sections []CompletenessSectionResponse `json:"sections"`
	Notice   string                        `json:"notice"`
}
type ComparisonResponse struct {
	BaseProfile       ProfileResponse    `json:"baseProfile"`
	ComparedProfile   ProfileResponse    `json:"comparedProfile"`
	InheritedEntities []SectionResponse  `json:"inheritedEntities"`
	HiddenEntities    []OverrideResponse `json:"hiddenEntities"`
	ModifiedFields    []OverrideResponse `json:"modifiedFields"`
	AppendedEntities  []SectionResponse  `json:"appendedEntities"`
}

type CreateEvidenceRequest struct {
	EntityType   string     `json:"entityType" validate:"required"`
	EntityID     uuid.UUID  `json:"entityId" validate:"required"`
	FieldName    *string    `json:"fieldName"`
	EvidenceType string     `json:"evidenceType" validate:"required"`
	DocumentID   *uuid.UUID `json:"documentId"`
	ImportID     *uuid.UUID `json:"importId"`
	SourceURL    *string    `json:"sourceUrl" validate:"omitempty,url"`
	SourceText   *string    `json:"sourceText"`
	Confidence   *float64   `json:"confidence" validate:"omitempty,gte=0,lte=1"`
}
type EvidenceResponse struct {
	ID           uuid.UUID  `json:"id"`
	ProfileID    uuid.UUID  `json:"profileId"`
	EntityType   string     `json:"entityType"`
	EntityID     uuid.UUID  `json:"entityId"`
	FieldName    *string    `json:"fieldName,omitempty"`
	EvidenceType string     `json:"evidenceType"`
	DocumentID   *uuid.UUID `json:"documentId,omitempty"`
	ImportID     *uuid.UUID `json:"importId,omitempty"`
	SourceURL    *string    `json:"sourceUrl,omitempty"`
	SourceText   *string    `json:"sourceText,omitempty"`
	Confidence   *float64   `json:"confidence,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}
