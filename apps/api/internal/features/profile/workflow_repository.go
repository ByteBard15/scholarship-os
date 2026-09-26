package profile

import (
	"context"

	"github.com/google/uuid"
)

type WorkflowRepository interface {
	WithinTransaction(context.Context, func(Repository, SectionRepository, WorkflowRepository) error) error

	CreateOverride(context.Context, *ProfileOverride) error
	GetOverride(context.Context, uuid.UUID, uuid.UUID) (*ProfileOverride, error)
	ListOverrides(context.Context, uuid.UUID) ([]ProfileOverride, error)
	UpdateOverride(context.Context, *ProfileOverride) error
	DeleteOverride(context.Context, uuid.UUID, uuid.UUID) error

	CreateSnapshot(context.Context, *ProfileSnapshot) error
	GetSnapshot(context.Context, uuid.UUID, uuid.UUID) (*ProfileSnapshot, error)
	ListSnapshots(context.Context, uuid.UUID) ([]ProfileSnapshot, error)

	CreateDocument(context.Context, *ProfileDocument) error
	GetDocument(context.Context, uuid.UUID, uuid.UUID) (*ProfileDocument, error)
	ListDocuments(context.Context, uuid.UUID) ([]ProfileDocument, error)
	DeleteDocument(context.Context, uuid.UUID, uuid.UUID) error

	CreateImport(context.Context, *ProfileImport) error
	GetImport(context.Context, uuid.UUID, uuid.UUID) (*ProfileImport, error)
	ListImports(context.Context, uuid.UUID) ([]ProfileImport, error)
	UpdateImport(context.Context, *ProfileImport) error
	CreateCandidates(context.Context, []ProfileImportCandidate) error
	GetCandidate(context.Context, uuid.UUID, uuid.UUID) (*ProfileImportCandidate, error)
	ListCandidates(context.Context, uuid.UUID) ([]ProfileImportCandidate, error)
	UpdateCandidate(context.Context, *ProfileImportCandidate) error
	ApplyImport(context.Context, *ProfileImport, []any, []ProfileEvidence, *AuditLog) error

	CreateEvidence(context.Context, *ProfileEvidence) error
	ListEvidence(context.Context, uuid.UUID) ([]ProfileEvidence, error)
	CreateAudit(context.Context, *AuditLog) error
}
