package workflow

import (
	"context"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
)

type ResearchPersistence struct {
	Run             *application.ResearchRun
	Sources         []application.ResearchSource
	Findings        []application.ResearchFinding
	Proposals       []ApplicationProposal
	ProposalSources []ApplicationProposalSource
	Outputs         []ResearchTaskOutput
	FollowUps       []ResearchTask
	Activities      []AgentActivity
}

type PrefillPersistence struct {
	Run                 *ApplicationPrefillRun
	Fields              []ApplicationField
	Answers             []ApplicationAnswer
	InformationRequests []InformationRequest
	Actions             []PrefillAction
	Tasks               []application.ApplicationTask
	Activities          []AgentActivity
}

type Repository interface {
	CreateResearchTask(context.Context, *ResearchTask, []ResearchTaskLink, []ResearchContext, []uuid.UUID) error
	ListResearchTasks(context.Context, ResearchTaskFilters) ([]ResearchTask, error)
	GetResearchTask(context.Context, uuid.UUID) (*ResearchTask, error)
	UpdateResearchTask(context.Context, *ResearchTask) error
	DeleteResearchTask(context.Context, *ResearchTask) error
	ListTaskLinks(context.Context, uuid.UUID) ([]ResearchTaskLink, error)
	GetTaskLink(context.Context, uuid.UUID, uuid.UUID) (*ResearchTaskLink, error)
	CreateTaskLink(context.Context, *ResearchTaskLink) error
	UpdateTaskLink(context.Context, *ResearchTaskLink) error
	DeleteTaskLink(context.Context, *ResearchTaskLink) error
	ListResearchContexts(context.Context, *uuid.UUID) ([]ResearchContext, error)
	GetResearchContext(context.Context, uuid.UUID) (*ResearchContext, error)
	CreateResearchContext(context.Context, *ResearchContext) error
	UpdateResearchContext(context.Context, *ResearchContext) error
	DeleteResearchContext(context.Context, *ResearchContext) error
	ListTaskContexts(context.Context, uuid.UUID) ([]ResearchContext, error)
	AttachTaskContexts(context.Context, uuid.UUID, []uuid.UUID) error
	DetachTaskContext(context.Context, uuid.UUID, uuid.UUID) error
	ListTaskOutputs(context.Context, uuid.UUID) ([]ResearchTaskOutput, error)
	ListTaskRuns(context.Context, uuid.UUID) ([]application.ResearchRun, error)
	ListRunSources(context.Context, uuid.UUID) ([]application.ResearchSource, error)
	ListRunFindings(context.Context, uuid.UUID) ([]application.ResearchFinding, error)
	ClaimResearchTask(context.Context, uuid.UUID, time.Time) (*ResearchTask, *application.ResearchRun, error)
	CompleteResearchTask(context.Context, uuid.UUID, uuid.UUID, time.Time) (*ResearchTask, error)
	PersistResearchTaskResult(context.Context, *ResearchTask, ResearchPersistence) error
	CreateAgentSource(context.Context, *application.ResearchSource, *AgentActivity) error
	CreateAgentFinding(context.Context, *application.ResearchFinding, *AgentActivity) error
	CreateAgentProposal(context.Context, *ResearchTask, *ApplicationProposal, []ApplicationProposalSource, *ResearchTaskOutput, *AgentActivity) error
	FailResearchTask(context.Context, *ResearchTask, *application.ResearchRun, error) error

	ListProposals(context.Context, *uuid.UUID, *string) ([]ApplicationProposal, error)
	GetProposal(context.Context, uuid.UUID) (*ApplicationProposal, error)
	ListProposalSources(context.Context, uuid.UUID) ([]ProposalSourceRecord, error)
	UpdateProposal(context.Context, *ApplicationProposal, *ResearchTask, AgentActivity) error
	ApproveProposal(context.Context, *ApplicationProposal, *ResearchTask, *profile.ApplicantProfile) (*application.Application, error)

	ListFields(context.Context, uuid.UUID) ([]ApplicationField, error)
	GetField(context.Context, uuid.UUID, uuid.UUID) (*ApplicationField, error)
	CreateField(context.Context, *ApplicationField) error
	UpdateField(context.Context, *ApplicationField) error
	DeleteField(context.Context, *ApplicationField) error
	ListQuestionnaires(context.Context, uuid.UUID) ([]ApplicationQuestionnaire, error)
	GetQuestionnaire(context.Context, uuid.UUID, uuid.UUID) (*ApplicationQuestionnaire, error)
	GetQuestionnaireByID(context.Context, uuid.UUID) (*ApplicationQuestionnaire, error)
	CreateQuestionnaire(context.Context, *ApplicationQuestionnaire) error
	UpdateQuestionnaire(context.Context, *ApplicationQuestionnaire) error
	DeleteQuestionnaire(context.Context, *ApplicationQuestionnaire) error
	ListQuestions(context.Context, uuid.UUID) ([]ApplicationQuestion, error)
	GetQuestion(context.Context, uuid.UUID, uuid.UUID) (*ApplicationQuestion, error)
	GetQuestionByID(context.Context, uuid.UUID) (*ApplicationQuestion, error)
	CreateQuestion(context.Context, *ApplicationQuestion) error
	UpdateQuestion(context.Context, *ApplicationQuestion) error
	DeleteQuestion(context.Context, *ApplicationQuestion) error
	ListAnswers(context.Context, uuid.UUID) ([]ApplicationAnswer, error)
	GetAnswer(context.Context, uuid.UUID, uuid.UUID) (*ApplicationAnswer, error)
	CreateAnswer(context.Context, *ApplicationAnswer) error
	UpdateAnswer(context.Context, *ApplicationAnswer) error
	RefreshQuestionnaireStatus(context.Context, uuid.UUID) error

	ListInformationRequests(context.Context, InformationRequestFilters) ([]InformationRequest, error)
	GetInformationRequest(context.Context, uuid.UUID) (*InformationRequest, error)
	ListInformationResponses(context.Context, uuid.UUID) ([]InformationRequestResponse, error)
	FindOpenInformationRequest(context.Context, InformationRequest) (*InformationRequest, error)
	CreateInformationRequest(context.Context, *InformationRequest) error
	RespondInformationRequest(context.Context, *InformationRequest, *InformationRequestResponse) error
	UpdateInformationRequest(context.Context, *InformationRequest) error
	ProcessInformationRequest(context.Context, *InformationRequest) error

	LoadPrefillInput(context.Context, uuid.UUID) (*PrefillData, error)
	PersistPrefillResult(context.Context, PrefillPersistence) error
	PreparationSummary(context.Context, uuid.UUID) (PreparationSummary, error)
}

type ProposalSourceRecord struct {
	Link   ApplicationProposalSource
	Source application.ResearchSource
}

type PrefillData struct {
	Application    *application.Application
	Requirements   []application.ApplicationRequirement
	Fields         []ApplicationField
	Questionnaires []QuestionnaireBundle
	CompletedInfo  []InformationRequest
	Findings       []application.ResearchFinding
}
