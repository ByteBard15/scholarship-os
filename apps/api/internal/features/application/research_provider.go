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

// MockApplicationResearcher is deterministic and performs no network or AI calls.
// Its values are proposals and always require human review.
type MockApplicationResearcher struct{ now func() time.Time }

func NewMockApplicationResearcher() *MockApplicationResearcher {
	return &MockApplicationResearcher{func() time.Time { return time.Now().UTC() }}
}
func (*MockApplicationResearcher) ProviderName() string { return "mock" }
func (*MockApplicationResearcher) ModelName() string    { return "deterministic-v1" }
func (m *MockApplicationResearcher) Research(_ context.Context, in ResearchInput) (*ResearchResult, error) {
	title := "Mock official opportunity page"
	publisher := "Example Provider"
	rawDeadline := "Applications close 31 January 2027"
	deadline := time.Date(2027, time.January, 31, 23, 59, 0, 0, time.UTC)
	confidence := 0.95
	mandatory := true
	official := true
	return &ResearchResult{Sources: []ResearchSourceCandidate{{Key: "official", URL: "https://example.test/opportunities/application", Title: &title, Publisher: &publisher, SourceType: "scholarship_provider", IsOfficial: true, RetrievedAt: m.now()}}, Requirements: []RequirementCandidate{{SourceKey: "official", Confidence: &confidence, RawText: ptrString("Applicants must provide an academic transcript."), VerificationStatus: "supported", Data: RequirementRequest{Category: "transcript", Title: "Academic transcript", IsMandatory: &mandatory, SourceURL: ptrString("https://example.test/opportunities/application")}}}, Deadlines: []DeadlineCandidate{{SourceKey: "official", Confidence: &confidence, RawText: &rawDeadline, VerificationStatus: "verified", Data: DeadlineRequest{DeadlineType: "scholarship", Title: "Scholarship application deadline", DeadlineAt: &deadline, Timezone: ptrString("UTC"), DatePrecision: "exact", RawDeadlineText: &rawDeadline, IsHardDeadline: &official, SourceURL: ptrString("https://example.test/opportunities/application"), VerifiedAt: ptrTime(m.now())}}}, Funding: []FundingCandidate{{SourceKey: "official", Confidence: &confidence, RawText: ptrString("Tuition fees are covered."), VerificationStatus: "supported", Data: FundingRequest{FundingType: "tuition", TuitionCoverage: ptrString("Full tuition coverage"), SourceURL: ptrString("https://example.test/opportunities/application")}}}}, nil
}
func ptrString(v string) *string     { return &v }
func ptrTime(v time.Time) *time.Time { return &v }
