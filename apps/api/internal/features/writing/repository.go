package writing

import (
	"context"

	"github.com/google/uuid"
)

type SampleFilters struct {
	DocumentType  *string
	ApplicationID *uuid.UUID
	Statuses      []string
}

type Repository interface {
	CreateSample(context.Context, *WritingSample, []WritingTag, []uuid.UUID) error
	GetSample(context.Context, uuid.UUID) (*WritingSample, error)
	ListSamples(context.Context, uuid.UUID, SampleFilters) ([]WritingSample, error)
	UpdateSample(context.Context, *WritingSample, *[]WritingTag) error
	DeleteSample(context.Context, *WritingSample) error
	ListSourceIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)
	CreateTag(context.Context, *WritingTag) error
	ListTags(context.Context, uuid.UUID) ([]TagResponse, error)
	GetTag(context.Context, uuid.UUID) (*WritingTag, error)
	DeleteTag(context.Context, *WritingTag) error
	ApplicationBelongsToUser(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	SamplesBelongToUser(context.Context, uuid.UUID, []uuid.UUID) (bool, error)
}
