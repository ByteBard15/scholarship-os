package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type listingRepository struct {
	Repository
	countFilters ResearchTaskFilters
	listFilters  ResearchTaskFilters
	total        int64
	tasks        []ResearchTask
}

func (r *listingRepository) CountResearchTasks(_ context.Context, filters ResearchTaskFilters) (int64, error) {
	r.countFilters = filters
	return r.total, nil
}
func (r *listingRepository) ListResearchTasks(_ context.Context, filters ResearchTaskFilters) ([]ResearchTask, error) {
	r.listFilters = filters
	return r.tasks, nil
}

func TestResearchPaginationKeepsCountsAndRowsInWorkspace(t *testing.T) {
	owner, other := uuid.New(), uuid.New()
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &owner})
	repo := &listingRepository{total: 41, tasks: []ResearchTask{{UserID: owner, Title: "Owned research"}}}
	service := NewService(repo, nil, nil, nil, nil)
	priority := "high"
	items, total, err := service.ListResearchTaskPage(ctx, ResearchTaskFilters{UserID: &other, Priority: &priority, SortBy: "dueAt"}, 3, 20)
	if err != nil || len(items) != 1 || total != 41 {
		t.Fatalf("page: items=%d total=%d err=%v", len(items), total, err)
	}
	for _, filters := range []ResearchTaskFilters{repo.countFilters, repo.listFilters} {
		if filters.UserID == nil || *filters.UserID != owner || filters.Priority == nil || *filters.Priority != priority {
			t.Fatalf("count/list escaped owner or filter: %+v", filters)
		}
	}
	if repo.listFilters.Limit != 20 || repo.listFilters.Offset != 40 || repo.listFilters.SortBy != "dueAt" {
		t.Fatalf("incorrect page: %+v", repo.listFilters)
	}
}

func TestResearchListingMetadataAndUnknownDeadline(t *testing.T) {
	repo := &listingRepository{total: 41, tasks: []ResearchTask{{Title: "No deadline"}}}
	handler := NewHandler(NewService(repo, nil, nil, nil, nil), validator.New())
	w := httptest.NewRecorder()
	handler.ListResearchTasks(w, httptest.NewRequest(http.MethodGet, "/?page=3&pageSize=20&sortBy=priority", nil))
	var result struct {
		Data []map[string]any                                       `json:"data"`
		Meta struct{ Count, Total, Page, PageSize, TotalPages int } `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || result.Meta.Count != 1 || result.Meta.Total != 41 || result.Meta.Page != 3 || result.Meta.PageSize != 20 || result.Meta.TotalPages != 3 {
		t.Fatalf("unexpected listing: %s", w.Body.String())
	}
	deadline, exists := result.Data[0]["dueAt"]
	if !exists || deadline != nil || result.Data[0]["priority"] != "normal" {
		t.Fatalf("unknown deadline/default priority: %+v", result.Data[0])
	}
}

func TestResearchListingRejectsInvalidPaginationAndSort(t *testing.T) {
	handler := NewHandler(nil, validator.New())
	for _, query := range []string{"page=0", "page=-1", "page=abc", "page=1000001", "pageSize=0", "pageSize=101", "pageSize=", "sortBy=invalid", "parentTaskId=invalid"} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ListResearchTasks(w, httptest.NewRequest(http.MethodGet, "/?"+query, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestResearchDeadlineCanBeClearedWithoutChangingPriority(t *testing.T) {
	repo := newFakeRepository()
	owner, taskID := uuid.New(), uuid.New()
	priority := "urgent"
	due := newTestService(repo, owner, nil, nil).now()
	repo.tasks[taskID] = &ResearchTask{Base: Base{ID: taskID}, UserID: owner, Title: "Research", Status: "draft", Priority: &priority, DueAt: &due}
	service := newTestService(repo, owner, nil, nil)
	task, err := service.UpdateResearchTask(context.Background(), taskID, UpdateResearchTaskRequest{ClearDueAt: true})
	if err != nil {
		t.Fatal(err)
	}
	if task.DueAt != nil || task.Priority == nil || *task.Priority != "urgent" {
		t.Fatalf("unexpected scheduling update: %+v", task)
	}
}

func TestAgentResearchQueueUsesPriorityAndDeadlineOrder(t *testing.T) {
	repo := &listingRepository{}
	handler := NewHandler(NewService(repo, nil, nil, nil, nil), validator.New())
	w := httptest.NewRecorder()
	handler.PollResearchTasks(w, httptest.NewRequest(http.MethodGet, "/?limit=5", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("queue returned %d: %s", w.Code, w.Body.String())
	}
	filters := repo.listFilters
	if filters.Status == nil || *filters.Status != "queued" || filters.Limit != 5 || filters.SortBy != "priority" || !filters.OldestFirst {
		t.Fatalf("unexpected queue ordering: %+v", filters)
	}
}
