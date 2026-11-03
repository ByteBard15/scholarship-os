package workflow

import (
	"context"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type queueRepository struct {
	*fakeRepository
	queueCalls int
	race       bool
}

func (r *queueRepository) QueueResearchTask(_ context.Context, id, owner uuid.UUID) (*ResearchTask, error) {
	r.queueCalls++
	task := r.tasks[id]
	if task == nil || task.UserID != owner {
		return nil, ErrResearchTaskNotFound
	}
	if r.race {
		task.Status = "running"
	}
	if task.Status != "draft" && task.Status != "ready" && task.Status != "queued" {
		return nil, ErrInvalidResearchTaskState
	}
	task.Status = "queued"
	return task, nil
}
func researchAgentContext(id uuid.UUID, scoped bool) context.Context {
	scopes := map[string]struct{}{}
	if scoped {
		scopes[principal.ScopeResearch] = struct{}{}
	}
	return principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, UserID: &id, Scopes: scopes})
}
func TestAgentQueueExistingResearch(t *testing.T) {
	owner, pid := uuid.New(), uuid.New()
	repo := &queueRepository{fakeRepository: newFakeRepository()}
	s := NewService(repo, nil, nil, &fakeProfiles{item: &profile.ApplicantProfile{Base: profile.Base{ID: pid}, UserID: owner}}, nil)
	ctx := researchAgentContext(owner, true)
	for _, state := range []string{"draft", "ready", "queued", "running", "completed", "failed", "cancelled"} {
		id := uuid.New()
		repo.tasks[id] = &ResearchTask{Base: Base{ID: id}, UserID: owner, ProfileID: &pid, Status: state}
		before := repo.queueCalls
		result, err := s.AgentQueueResearchTask(ctx, id)
		if state == "draft" || state == "ready" || state == "queued" {
			if err != nil || result.Status != "queued" {
				t.Fatalf("%s: %+v %v", state, result, err)
			}
			if state == "queued" && repo.queueCalls != before {
				t.Fatal("idempotent queue rewrote state")
			}
		} else if err != ErrInvalidResearchTaskState || repo.queueCalls != before {
			t.Fatalf("%s was requeued: %v", state, err)
		}
	}
	id := uuid.New()
	repo.tasks[id] = &ResearchTask{Base: Base{ID: id}, UserID: owner, Status: "draft", ProfileID: &pid}
	for _, test := range []struct {
		ctx  context.Context
		want error
	}{
		{researchAgentContext(uuid.New(), true), ErrResearchTaskNotFound},
		{researchAgentContext(owner, false), principal.ErrAgentScopeRequired},
		{context.Background(), principal.ErrAuthRequired},
		{principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &owner}), principal.ErrAgentRequired},
	} {
		before := repo.queueCalls
		if _, err := s.AgentQueueResearchTask(test.ctx, id); err != test.want || repo.queueCalls != before {
			t.Fatalf("guard: %v", err)
		}
	}
	repo.tasks[id].ProfileID = nil
	if _, err := s.AgentQueueResearchTask(ctx, id); err == nil {
		t.Fatal("missing profile accepted")
	}
	foreign := uuid.New()
	repo.tasks[id].ProfileID = &foreign
	if _, err := s.AgentQueueResearchTask(ctx, id); err == nil {
		t.Fatal("unowned profile accepted")
	}
	repo.tasks[id].ProfileID = &pid
	repo.race = true
	if _, err := s.AgentQueueResearchTask(ctx, id); err != ErrInvalidResearchTaskState || repo.tasks[id].Status != "running" {
		t.Fatal("concurrent claim overwritten", err)
	}
}
func TestAgentListDraftTasksUsesWorkspaceAndPagination(t *testing.T) {
	owner, other := uuid.New(), uuid.New()
	draft := "draft"
	repo := &listingRepository{total: 2, tasks: []ResearchTask{{UserID: owner}}}
	s := NewService(repo, nil, nil, nil, nil)
	_, total, err := s.AgentListResearchTaskPage(researchAgentContext(owner, true), ResearchTaskFilters{UserID: &other, Status: &draft}, 2, 50)
	if err != nil || total != 2 || repo.listFilters.UserID == nil || *repo.listFilters.UserID != owner || repo.listFilters.Offset != 50 || repo.listFilters.Status != &draft {
		t.Fatalf("list guard: %+v %v", repo.listFilters, err)
	}
	if _, _, err := s.AgentListResearchTaskPage(researchAgentContext(owner, false), ResearchTaskFilters{}, 1, 20); err != principal.ErrAgentScopeRequired {
		t.Fatal(err)
	}
}
