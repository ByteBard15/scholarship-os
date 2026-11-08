package workflow

import (
	"context"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
	"strings"
	"testing"
)

type deletionRepo struct {
	*fakeRepository
	calls int
}

func (r *deletionRepo) PreviewResearchDeletion(context.Context, *ResearchTask) (*ResearchDeletionPreview, error) {
	r.calls++
	return &ResearchDeletionPreview{}, nil
}
func (r *deletionRepo) CascadeDeleteResearch(context.Context, *ResearchTask, string) error {
	r.calls++
	return nil
}
func TestResearchDeletionAuthorization(t *testing.T) {
	owner, other, id := uuid.New(), uuid.New(), uuid.New()
	repo := &deletionRepo{fakeRepository: newFakeRepository()}
	repo.tasks[id] = &ResearchTask{UserID: owner}
	repo.tasks[id].ID = id
	service := NewService(repo, nil, nil, nil, nil)
	for _, tc := range []struct {
		name    string
		actor   *principal.Principal
		allowed bool
	}{
		{"anonymous", nil, false},
		{"owner", &principal.Principal{Type: principal.PrincipalUser, UserID: &owner}, true},
		{"foreign owner", &principal.Principal{Type: principal.PrincipalUser, UserID: &other}, false},
		{"agent", &principal.Principal{Type: principal.PrincipalAgent, UserID: &owner}, false},
		{"preparation scopes without delete", &principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeResearch: {}, principal.ScopeApplications: {}, principal.ScopeTasks: {}}}, false},
		{"delete agent", &principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeDelete: {}}}, true},
		{"foreign delete agent", &principal.Principal{Type: principal.PrincipalAgent, UserID: &other, Scopes: map[string]struct{}{principal.ScopeDelete: {}}}, false},
		{"unbound delete agent", &principal.Principal{Type: principal.PrincipalAgent, Scopes: map[string]struct{}{principal.ScopeDelete: {}}}, false},
		{"system", &principal.Principal{Type: principal.PrincipalSystem}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.actor != nil {
				ctx = principal.WithPrincipal(ctx, *tc.actor)
			}
			before := repo.calls
			_, err := service.PreviewResearchDeletion(ctx, id)
			if (err == nil) != tc.allowed {
				t.Fatalf("preview error %v", err)
			}
			err = service.CascadeDeleteResearch(ctx, id, strings.Repeat("a", 64))
			if (err == nil) != tc.allowed {
				t.Fatalf("delete error %v", err)
			}
			if !tc.allowed && repo.calls != before {
				t.Fatal("unauthorized persistence access")
			}
		})
	}
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser, UserID: &owner})
	before := repo.calls
	if err := service.CascadeDeleteResearch(ctx, id, ""); err == nil || repo.calls != before {
		t.Fatal("missing confirmation reached repository")
	}
}
