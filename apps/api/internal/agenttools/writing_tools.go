package agenttools

import (
	"context"

	"github.com/byte/scholarship-os/apps/api/internal/features/writing"
)

// WritingAgentTools exposes a narrow retrieval and generation surface. Agents
// discover available tags before loading matching source content and may only
// save generated work as a reviewable suggested sample.
type WritingAgentTools interface {
	ListWritingTags(context.Context) ([]writing.TagResponse, error)
	MatchWritingSamples(context.Context, writing.MatchSamplesRequest) ([]writing.SampleMatchResponse, error)
	SaveGeneratedWriting(context.Context, writing.CreateSampleRequest) (*writing.SampleResponse, error)
}

// ControlledWritingTools intentionally contains no repository or database
// access. The service applies ownership, provenance, and review-state rules.
type ControlledWritingTools struct{ writings *writing.Service }

func NewWritingTools(writings *writing.Service) *ControlledWritingTools {
	return &ControlledWritingTools{writings: writings}
}

func (t *ControlledWritingTools) ListWritingTags(ctx context.Context) ([]writing.TagResponse, error) {
	return t.writings.ListTags(ctx, nil)
}

func (t *ControlledWritingTools) MatchWritingSamples(ctx context.Context, request writing.MatchSamplesRequest) ([]writing.SampleMatchResponse, error) {
	return t.writings.MatchSamples(ctx, request)
}

func (t *ControlledWritingTools) SaveGeneratedWriting(ctx context.Context, request writing.CreateSampleRequest) (*writing.SampleResponse, error) {
	return t.writings.CreateSample(ctx, request, true)
}
