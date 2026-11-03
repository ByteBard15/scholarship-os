package workflow

import (
	"context"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/google/uuid"
	"net/http"
)

func requireResearchAgent(ctx context.Context) error {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return err
	}
	if actor.IsSystem() {
		return nil
	}
	if actor.Type != principal.PrincipalAgent {
		return principal.ErrAgentRequired
	}
	if !actor.HasScope(principal.ScopeResearch) {
		return principal.ErrAgentScopeRequired
	}
	if actor.UserID == nil {
		return principal.ErrForbidden
	}
	return nil
}

func (s *Service) AgentListResearchTaskPage(ctx context.Context, filters ResearchTaskFilters, page, pageSize int) ([]ResearchTask, int64, error) {
	if err := requireResearchAgent(ctx); err != nil {
		return nil, 0, err
	}
	return s.ListResearchTaskPage(ctx, filters, page, pageSize)
}

func (s *Service) AgentQueueResearchTask(ctx context.Context, id uuid.UUID) (*ResearchTask, error) {
	if err := requireResearchAgent(ctx); err != nil {
		return nil, err
	}
	task, err := s.GetResearchTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if task.Status == "queued" {
		return task, nil
	}
	if task.Status != "draft" && task.Status != "ready" {
		return nil, ErrInvalidResearchTaskState
	}
	if task.ProfileID == nil {
		return nil, validationError("profileId is required before a research task can be queued")
	}
	if err := s.validateTaskRelationships(ctx, task); err != nil {
		return nil, err
	}
	result, err := s.repo.QueueResearchTask(ctx, task.ID, task.UserID)
	if err == nil {
		appLogger.Info(ctx, "agent queued research task", "research_task_id", task.ID)
	}
	return result, err
}
func (h *Handler) AgentQueueResearchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	task, err := h.service.AgentQueueResearchTask(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, researchTaskResponse(task))
}
