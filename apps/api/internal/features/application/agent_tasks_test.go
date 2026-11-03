package application

import (
	"context"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
	"testing"
)

func TestAgentChecklistEditsStayInWorkspace(t *testing.T) {
	owner, appID, taskID := uuid.New(), uuid.New(), uuid.New()
	r := &fakeApplicationRepo{app: &Application{Base: Base{ID: appID}, UserID: owner}, tasks: []ApplicationTask{{Base: Base{ID: taskID}, ApplicationID: appID, Title: "Review CV", Status: "todo"}}}
	s := NewService(r, nil, nil)
	actor := principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeTasks: {}}}
	ctx := principal.WithPrincipal(context.Background(), actor)
	updated, err := s.UpdateTask(ctx, appID, taskID, TaskRequest{Title: "Review biomedical CV"})
	if err != nil || updated.Title != "Review biomedical CV" {
		t.Fatal(err)
	}
	actor.Scopes = map[string]struct{}{}
	if _, err := s.UpdateTask(principal.WithPrincipal(context.Background(), actor), appID, taskID, TaskRequest{Title: "No"}); err != principal.ErrAgentScopeRequired {
		t.Fatal("unscoped task edit", err)
	}
	other := uuid.New()
	actor.UserID = &other
	actor.Scopes = map[string]struct{}{principal.ScopeTasks: {}}
	if _, err := s.UpdateTask(principal.WithPrincipal(context.Background(), actor), appID, taskID, TaskRequest{Title: "No"}); err != ErrApplicationNotFound {
		t.Fatal("foreign task edit", err)
	}
}
