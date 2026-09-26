package workflow

import (
	"context"
	"strings"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
)

type ResearchTaskInput struct {
	Task              ResearchTaskResponse
	Links             []TaskLinkResponse
	TargetApplication *application.ApplicationResponse
	EffectiveProfile  *profile.EffectiveProfileResponse
}

type QuestionnaireCandidate struct {
	StableKey         string            `json:"stableKey"`
	Title             string            `json:"title"`
	Description       *string           `json:"description,omitempty"`
	QuestionnaireType *string           `json:"questionnaireType,omitempty"`
	SourceURL         *string           `json:"sourceUrl,omitempty"`
	Questions         []QuestionRequest `json:"questions"`
}

type ResearchTaskCandidate struct {
	StableKey    string  `json:"stableKey"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
	TaskType     string  `json:"taskType"`
	Priority     *string `json:"priority,omitempty"`
}

type ResearchTaskResult struct {
	Sources              []application.ResearchSourceCandidate
	Findings             []application.FindingCandidate
	ApplicationProposals []ApplicationProposalCandidate
	Requirements         []application.RequirementCandidate
	Deadlines            []application.DeadlineCandidate
	Funding              []application.FundingCandidate
	Contacts             []application.ContactCandidate
	Supervisors          []application.SupervisorCandidate
	URLs                 []application.URLCandidate
	Questionnaires       []QuestionnaireCandidate
	SuggestedTasks       []ResearchTaskCandidate
}

type ResearchExecutor interface {
	Execute(context.Context, ResearchTaskInput) (*ResearchTaskResult, error)
	ProviderName() string
	ModelName() string
}

type MockWebResearchProvider struct{ now func() time.Time }

func NewMockWebResearchProvider() *MockWebResearchProvider {
	return &MockWebResearchProvider{now: func() time.Time { return time.Now().UTC() }}
}

func (*MockWebResearchProvider) ProviderName() string { return "mock" }
func (*MockWebResearchProvider) ModelName() string    { return "deterministic-research-v1" }

func (m *MockWebResearchProvider) Execute(_ context.Context, input ResearchTaskInput) (*ResearchTaskResult, error) {
	title := "Example Graduate Scholarship official page"
	publisher := "Example University"
	country := "Exampleland"
	degree := "masters"
	scholarshipType := "full"
	officialURL := "https://example.test/scholarships/graduate-2027"
	programmeURL := "https://example.test/programmes/biomedical-engineering"
	confidence := 0.96
	intake := "Autumn 2027"
	intakeYear := 2027
	summary := "A fictional fully funded graduate opportunity returned by the deterministic local research provider."
	reasoning := "The proposal is supported by the linked official scholarship and programme pages."
	deadline := time.Date(2027, time.January, 31, 23, 59, 0, 0, time.UTC)
	rawDeadline := "Applications close 31 January 2027"
	mandatory := true
	hard := true
	questionnaireType := "scholarship"
	link := officialURL
	if len(input.Links) > 0 {
		link = input.Links[0].URL
	}
	proposalName := strings.TrimSpace(input.Task.Title)
	if proposalName == "" {
		proposalName = "Example Graduate Scholarship 2027"
	}
	return &ResearchTaskResult{
		Sources:              []application.ResearchSourceCandidate{{Key: "official-scholarship", URL: link, Title: &title, Publisher: &publisher, SourceType: "scholarship_provider", IsOfficial: true, RetrievedAt: m.now()}},
		Findings:             []application.FindingCandidate{{SourceKey: "official-scholarship", Category: "identity", Field: "scholarship", Value: map[string]any{"name": "Example Graduate Scholarship", "country": country}, Confidence: &confidence, VerificationStatus: "verified"}},
		ApplicationProposals: []ApplicationProposalCandidate{{StableKey: "example-graduate-scholarship-2027", ProposedInstitution: &ProposedInstitution{Name: "Example University", InstitutionType: stringPointer("university"), Country: country, WebsiteURL: stringPointer("https://example.test/university")}, ProposedProgramme: &ProposedProgramme{Name: "MSc Biomedical Engineering", DegreeLevel: &degree, FieldOfStudy: stringPointer("Biomedical Engineering"), ProgrammeURL: &programmeURL}, ProposedScholarship: &ProposedScholarship{Name: "Example Graduate Scholarship", ProviderName: &publisher, Country: &country, DegreeLevel: &degree, ScholarshipType: &scholarshipType, OfficialURL: &officialURL}, Name: proposalName, Country: &country, Intake: &intake, IntakeYear: &intakeYear, Summary: &summary, Confidence: &confidence, ReasoningSummary: &reasoning, SourceKeys: []string{"official-scholarship"}}},
		Requirements:         []application.RequirementCandidate{{SourceKey: "official-scholarship", Data: application.RequirementRequest{Category: "transcript", Title: "Academic transcript", IsMandatory: &mandatory, SourceURL: &link}, Confidence: &confidence, VerificationStatus: "supported"}},
		Deadlines:            []application.DeadlineCandidate{{SourceKey: "official-scholarship", Data: application.DeadlineRequest{DeadlineType: "scholarship", Title: "Scholarship application deadline", DeadlineAt: &deadline, Timezone: stringPointer("UTC"), DatePrecision: "exact", RawDeadlineText: &rawDeadline, IsHardDeadline: &hard, SourceURL: &link, VerifiedAt: timePointer(m.now())}, Confidence: &confidence, RawText: &rawDeadline, VerificationStatus: "verified"}},
		Funding:              []application.FundingCandidate{{SourceKey: "official-scholarship", Data: application.FundingRequest{FundingType: "full", TuitionCoverage: stringPointer("Full tuition"), StipendPeriod: stringPointer("monthly"), SourceURL: &link}, Confidence: &confidence, VerificationStatus: "supported"}},
		URLs:                 []application.URLCandidate{{SourceKey: "official-scholarship", Data: application.URLRequest{URLType: "scholarship", Label: stringPointer("Official scholarship page"), URL: link, IsOfficial: true}, Confidence: &confidence, VerificationStatus: "verified"}},
		Questionnaires:       []QuestionnaireCandidate{{StableKey: "initial-scholarship-questionnaire", Title: "Initial scholarship questionnaire", QuestionnaireType: &questionnaireType, SourceURL: &link, Questions: []QuestionRequest{{Key: stringPointer("undergraduate_institution"), Prompt: "What is your undergraduate institution?", QuestionType: "short_text", IsRequired: true, SortOrder: intPointer(1)}, {Key: stringPointer("passport_expiry"), Prompt: "What is your passport expiry date?", QuestionType: "date", IsRequired: true, SortOrder: intPointer(2)}, {Key: stringPointer("programme_motivation"), Prompt: "Why are you interested in this programme?", QuestionType: "long_text", IsRequired: true, SortOrder: intPointer(3)}}}},
	}, nil
}

type PrefillInput struct {
	Application                  application.ApplicationResponse
	EffectiveProfile             *profile.EffectiveProfileResponse
	Requirements                 []application.ApplicationRequirement
	Fields                       []ApplicationField
	Questionnaires               []QuestionnaireBundle
	CompletedInformationRequests []InformationRequest
	ResearchFindings             []application.ResearchFinding
	Regenerate                   bool
}

type QuestionnaireBundle struct {
	Questionnaire ApplicationQuestionnaire
	Questions     []ApplicationQuestion
	Answers       map[uuid.UUID][]ApplicationAnswer
}

type FieldUpdateCandidate struct {
	FieldID        *uuid.UUID
	Key            string
	Label          string
	Value          any
	ValueType      string
	Status         string
	SourceType     string
	SourceEntityID *uuid.UUID
}

type QuestionAnswerCandidate struct {
	QuestionID       uuid.UUID
	Value            any
	DraftText        *string
	Status           string
	AnswerSource     string
	SourceEntityType *string
	SourceEntityID   *uuid.UUID
	Confidence       *float64
}

type InformationRequestCandidate struct {
	ApplicationID      uuid.UUID
	QuestionnaireID    *uuid.UUID
	QuestionID         *uuid.UUID
	ApplicationFieldID *uuid.UUID
	RequestType        string
	Title              string
	Prompt             string
	Context            *string
	ResponseType       string
	Priority           *string
}

type PrefillResult struct {
	FieldUpdates        []FieldUpdateCandidate
	QuestionAnswers     []QuestionAnswerCandidate
	InformationRequests []InformationRequestCandidate
	SuggestedTasks      []application.ApplicationTask
}

type ApplicationPrefiller interface {
	Prefill(context.Context, PrefillInput) (*PrefillResult, error)
	ProviderName() string
	ModelName() string
}

type MockApplicationPrefiller struct{}

func NewMockApplicationPrefiller() *MockApplicationPrefiller { return &MockApplicationPrefiller{} }
func (*MockApplicationPrefiller) ProviderName() string       { return "mock" }
func (*MockApplicationPrefiller) ModelName() string          { return "deterministic-prefill-v1" }

func (*MockApplicationPrefiller) Prefill(_ context.Context, input PrefillInput) (*PrefillResult, error) {
	result := &PrefillResult{}
	for _, bundle := range input.Questionnaires {
		for _, question := range bundle.Questions {
			if !input.Regenerate && hasProtectedAnswer(bundle.Answers[question.ID]) {
				continue
			}
			key := ""
			if question.Key != nil {
				key = *question.Key
			}
			switch key {
			case "undergraduate_institution":
				if input.EffectiveProfile != nil && len(input.EffectiveProfile.Education) > 0 {
					institution := input.EffectiveProfile.Education[0].Institution
					if institution != nil && *institution != "" {
						confidence := 1.0
						result.QuestionAnswers = append(result.QuestionAnswers, QuestionAnswerCandidate{QuestionID: question.ID, Value: *institution, Status: "answered", AnswerSource: "profile", SourceEntityType: stringPointer("education"), SourceEntityID: &input.EffectiveProfile.Education[0].ID, Confidence: &confidence})
						continue
					}
				}
				result.InformationRequests = append(result.InformationRequests, missingQuestion(input.Application.ID, bundle.Questionnaire.ID, question, "Please provide your undergraduate institution.", "text"))
			case "passport_expiry":
				result.InformationRequests = append(result.InformationRequests, missingQuestion(input.Application.ID, bundle.Questionnaire.ID, question, "What is your passport expiry date?", "date"))
			case "programme_motivation":
				draft := "I am interested in this programme because its academic focus aligns with my documented education, experience, and research goals. Please add your personal reasons before approval."
				confidence := 0.55
				result.QuestionAnswers = append(result.QuestionAnswers, QuestionAnswerCandidate{QuestionID: question.ID, DraftText: &draft, Status: "suggested", AnswerSource: "agent", Confidence: &confidence})
			default:
				result.InformationRequests = append(result.InformationRequests, missingQuestion(input.Application.ID, bundle.Questionnaire.ID, question, question.Prompt, responseTypeForQuestion(question.QuestionType)))
			}
		}
	}
	return result, nil
}

func hasProtectedAnswer(answers []ApplicationAnswer) bool {
	for _, answer := range answers {
		if answer.Status == "approved" || (answer.AnswerSource != nil && (*answer.AnswerSource == "user" || *answer.AnswerSource == "information_request")) {
			return true
		}
	}
	return false
}

func missingQuestion(applicationID, questionnaireID uuid.UUID, question ApplicationQuestion, prompt, responseType string) InformationRequestCandidate {
	return InformationRequestCandidate{ApplicationID: applicationID, QuestionnaireID: &questionnaireID, QuestionID: &question.ID, RequestType: "questionnaire_answer", Title: "Information needed for application question", Prompt: prompt, ResponseType: responseType, Priority: stringPointer("normal")}
}

func responseTypeForQuestion(questionType string) string {
	switch questionType {
	case "long_text":
		return "long_text"
	case "number", "date", "boolean", "url", "email":
		return questionType
	case "single_select":
		return "select"
	case "multi_select":
		return "multiselect"
	default:
		return "text"
	}
}

func stringPointer(value string) *string     { return &value }
func intPointer(value int) *int              { return &value }
func timePointer(value time.Time) *time.Time { return &value }
