package application

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ResearchInput struct {
	Application             ApplicationResponse `json:"application"`
	EffectiveProfileSummary string              `json:"effectiveProfileSummary,omitempty"`
}
type ResearchSourceCandidate struct {
	Key         string
	URL         string
	Title       *string
	Publisher   *string
	SourceType  string
	IsOfficial  bool
	RetrievedAt time.Time
	PublishedAt *time.Time
	ContentHash *string
	Notes       *string
}
type sourcedCandidate[T any] struct {
	SourceKey          string
	Data               T
	Confidence         *float64
	RawText            *string
	VerificationStatus string
}
type RequirementCandidate = sourcedCandidate[RequirementRequest]
type DeadlineCandidate = sourcedCandidate[DeadlineRequest]
type FundingCandidate = sourcedCandidate[FundingRequest]
type ContactCandidate = sourcedCandidate[ContactRequest]
type SupervisorCandidate = sourcedCandidate[SupervisorRequest]
type URLCandidate = sourcedCandidate[URLRequest]
type FindingCandidate struct {
	SourceKey          string
	Category           string
	Field              string
	Value              any
	Confidence         *float64
	RawText            *string
	VerificationStatus string
	Notes              *string
}
type FindingProposal struct {
	SourceID           *uuid.UUID
	Category           string
	Field              string
	Value              any
	Confidence         *float64
	RawText            *string
	VerificationStatus string
	Notes              *string
}
type ResearchResult struct {
	Sources      []ResearchSourceCandidate
	Requirements []RequirementCandidate
	Deadlines    []DeadlineCandidate
	Funding      []FundingCandidate
	Contacts     []ContactCandidate
	Supervisors  []SupervisorCandidate
	URLs         []URLCandidate
	Findings     []FindingCandidate
}
type ApplicationResearcher interface {
	Research(context.Context, ResearchInput) (*ResearchResult, error)
	ProviderName() string
	ModelName() string
}

func ptrString(v string) *string { return &v }
