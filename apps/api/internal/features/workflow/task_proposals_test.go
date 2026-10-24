package workflow

import (
	"context"
	"testing"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type taskProposalRepository struct {
	*fakeRepository
	items     []ApplicationProposal
	requested uuid.UUID
}

func (r *taskProposalRepository) ListTaskProposals(_ context.Context, taskID uuid.UUID) ([]ApplicationProposal, error) {
	r.requested = taskID
	return r.items, nil
}

func TestTaskProposalsKeepApprovedAndMaterializedProposals(t *testing.T) {
	owner, taskID, applicationID := uuid.New(), uuid.New(), uuid.New()
	repo := &taskProposalRepository{fakeRepository: newFakeRepository(), items: []ApplicationProposal{
		{Base: Base{ID: uuid.New()}, UserID: owner, ResearchTaskID: taskID, Status: "pending"},
		{Base: Base{ID: uuid.New()}, UserID: owner, ResearchTaskID: taskID, Status: "approved"},
		{Base: Base{ID: uuid.New()}, UserID: owner, ResearchTaskID: taskID, Status: "approved", ApplicationID: &applicationID},
		{Base: Base{ID: uuid.New()}, UserID: owner, ResearchTaskID: taskID, Status: "rejected"},
	}}
	repo.tasks[taskID] = &ResearchTask{Base: Base{ID: taskID}, UserID: owner}
	service := NewService(repo, nil, nil, nil, nil)
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &owner})
	items, err := service.ListTaskProposals(ctx, taskID)
	if err != nil || len(items) != 4 || repo.requested != taskID {
		t.Fatalf("lost reviewed proposals: %+v err=%v", items, err)
	}
	if items[1].Status != "approved" || items[2].ApplicationID == nil || *items[2].ApplicationID != applicationID {
		t.Fatalf("lost approval/application details: %+v", items)
	}
	other := uuid.New()
	ctx = principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &other})
	repo.requested = uuid.Nil
	if _, err := service.ListTaskProposals(ctx, taskID); err != ErrResearchTaskNotFound || repo.requested != uuid.Nil {
		t.Fatalf("foreign workspace proposal query allowed: err=%v", err)
	}
}
