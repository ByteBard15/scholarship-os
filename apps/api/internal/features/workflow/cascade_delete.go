package workflow

import (
	"context"
	"net/http"
	"strings"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/google/uuid"
)

type DeletionItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
type ResearchDeletionPreview struct {
	Tasks             []DeletionItem `json:"tasks"`
	Proposals         []DeletionItem `json:"proposals"`
	Applications      []DeletionItem `json:"applications"`
	ArchivedProfiles  []DeletionItem `json:"archivedProfiles"`
	Blockers          []string       `json:"blockers"`
	ConfirmationToken string         `json:"confirmationToken"`
}
type CascadeDeleteResearchRequest struct {
	ConfirmationToken string `json:"confirmationToken" validate:"required,len=64"`
}

func (s *Service) authorizeResearchDeletion(ctx context.Context, id uuid.UUID) (*ResearchTask, error) {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if actor.Type != principal.PrincipalUser && !actor.IsSystem() {
		return nil, principal.ErrForbidden
	}
	return s.GetResearchTask(ctx, id)
}
func (s *Service) PreviewResearchDeletion(ctx context.Context, id uuid.UUID) (*ResearchDeletionPreview, error) {
	task, err := s.authorizeResearchDeletion(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.repo.PreviewResearchDeletion(ctx, task)
}
func (s *Service) CascadeDeleteResearch(ctx context.Context, id uuid.UUID, token string) error {
	task, err := s.authorizeResearchDeletion(ctx, id)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(token)) != 64 {
		return validationError("refresh the deletion preview before confirming")
	}
	err = s.repo.CascadeDeleteResearch(ctx, task, token)
	if err == nil {
		appLogger.FromContext(ctx).Info("research task cascade deleted", "task_id", id)
	}
	return err
}
func (h *Handler) PreviewResearchDeletion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	result, err := h.service.PreviewResearchDeletion(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}
func (h *Handler) CascadeDeleteResearch(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var q CascadeDeleteResearchRequest
	if !h.decode(w, r, &q) {
		return
	}
	if err := h.service.CascadeDeleteResearch(r.Context(), id, q.ConfirmationToken); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
