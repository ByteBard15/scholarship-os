package agenttools

import (
	"context"

	"github.com/byte/scholarship-os/apps/api/internal/features/workflow"
	"github.com/google/uuid"
)

// WorkflowAgentTools is the narrow mutation surface available to research and
// prefill agents. It intentionally exposes domain operations rather than a
// generic entity update or database handle.
type WorkflowAgentTools interface {
	GetResearchTask(context.Context, uuid.UUID) (*workflow.ResearchTask, error)
	ListResearchTaskLinks(context.Context, uuid.UUID) ([]workflow.ResearchTaskLink, error)
	GetApplicationProposal(context.Context, uuid.UUID) (*workflow.ApplicationProposalResponse, error)
	CreateApplicationQuestionnaire(context.Context, uuid.UUID, workflow.QuestionnaireRequest) (*workflow.ApplicationQuestionnaire, error)
	CreateApplicationQuestion(context.Context, uuid.UUID, workflow.QuestionRequest) (*workflow.ApplicationQuestion, error)
	UpdateApplicationField(context.Context, uuid.UUID, uuid.UUID, workflow.ApplicationFieldRequest) (*workflow.ApplicationField, error)
	CreateInformationRequest(context.Context, workflow.CreateInformationRequest) (*workflow.InformationRequest, error)
	GetInformationRequest(context.Context, uuid.UUID) (*workflow.InformationRequestResponseDTO, error)
	CompleteInformationRequest(context.Context, uuid.UUID) (*workflow.InformationRequest, error)
	CompleteResearchTask(context.Context, uuid.UUID) (*workflow.ResearchTask, error)
}

type workflowService interface {
	GetResearchTask(context.Context, uuid.UUID) (*workflow.ResearchTask, error)
	ListTaskLinks(context.Context, uuid.UUID) ([]workflow.ResearchTaskLink, error)
	GetProposal(context.Context, uuid.UUID) (*workflow.ApplicationProposalResponse, error)
	CreateQuestionnaire(context.Context, uuid.UUID, workflow.QuestionnaireRequest) (*workflow.ApplicationQuestionnaire, error)
	CreateQuestion(context.Context, uuid.UUID, workflow.QuestionRequest) (*workflow.ApplicationQuestion, error)
	UpdateField(context.Context, uuid.UUID, uuid.UUID, workflow.ApplicationFieldRequest) (*workflow.ApplicationField, error)
	CreateInformationRequest(context.Context, workflow.CreateInformationRequest) (*workflow.InformationRequest, error)
	GetInformationRequest(context.Context, uuid.UUID) (*workflow.InformationRequestResponseDTO, error)
	ProcessInformationRequest(context.Context, uuid.UUID) (*workflow.InformationRequest, error)
	TransitionResearchTask(context.Context, uuid.UUID, string) (*workflow.ResearchTask, error)
}

type ControlledWorkflowTools struct{ workflows workflowService }

func NewWorkflowTools(workflows workflowService) *ControlledWorkflowTools {
	return &ControlledWorkflowTools{workflows: workflows}
}

func (t *ControlledWorkflowTools) GetResearchTask(ctx context.Context, id uuid.UUID) (*workflow.ResearchTask, error) {
	return t.workflows.GetResearchTask(ctx, id)
}
func (t *ControlledWorkflowTools) ListResearchTaskLinks(ctx context.Context, id uuid.UUID) ([]workflow.ResearchTaskLink, error) {
	return t.workflows.ListTaskLinks(ctx, id)
}
func (t *ControlledWorkflowTools) GetApplicationProposal(ctx context.Context, id uuid.UUID) (*workflow.ApplicationProposalResponse, error) {
	return t.workflows.GetProposal(ctx, id)
}
func (t *ControlledWorkflowTools) CreateApplicationQuestionnaire(ctx context.Context, applicationID uuid.UUID, request workflow.QuestionnaireRequest) (*workflow.ApplicationQuestionnaire, error) {
	return t.workflows.CreateQuestionnaire(ctx, applicationID, request)
}
func (t *ControlledWorkflowTools) CreateApplicationQuestion(ctx context.Context, questionnaireID uuid.UUID, request workflow.QuestionRequest) (*workflow.ApplicationQuestion, error) {
	return t.workflows.CreateQuestion(ctx, questionnaireID, request)
}
func (t *ControlledWorkflowTools) UpdateApplicationField(ctx context.Context, applicationID, fieldID uuid.UUID, request workflow.ApplicationFieldRequest) (*workflow.ApplicationField, error) {
	return t.workflows.UpdateField(ctx, applicationID, fieldID, request)
}
func (t *ControlledWorkflowTools) CreateInformationRequest(ctx context.Context, request workflow.CreateInformationRequest) (*workflow.InformationRequest, error) {
	return t.workflows.CreateInformationRequest(ctx, request)
}
func (t *ControlledWorkflowTools) GetInformationRequest(ctx context.Context, id uuid.UUID) (*workflow.InformationRequestResponseDTO, error) {
	return t.workflows.GetInformationRequest(ctx, id)
}
func (t *ControlledWorkflowTools) CompleteInformationRequest(ctx context.Context, id uuid.UUID) (*workflow.InformationRequest, error) {
	return t.workflows.ProcessInformationRequest(ctx, id)
}
func (t *ControlledWorkflowTools) CompleteResearchTask(ctx context.Context, id uuid.UUID) (*workflow.ResearchTask, error) {
	return t.workflows.TransitionResearchTask(ctx, id, "complete")
}
