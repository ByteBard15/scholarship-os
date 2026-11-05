package application

import (
	"context"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
	"testing"
)

func TestValidateTaskEvidence(t *testing.T) {
	for _, q := range []TaskEvidenceRequest{
		{Kind: "identifier", Label: "UCID", Value: "example-not-an-issued-id"},
		{Kind: "link", Label: "Portal", Value: "https://example.edu/application"},
		{Kind: "confirmation", Label: "Saved draft", Value: "Observed in portal; not submitted"},
	} {
		if err := validateTaskEvidence(q); err != nil {
			t.Fatal(err)
		}
	}
	for _, q := range []TaskEvidenceRequest{
		{Kind: "link", Label: "Unsafe", Value: "javascript:alert(1)"},
		{Kind: "link", Label: "Credential URL", Value: "https://user:secret@example.edu"},
		{Kind: "identifier", Label: " ", Value: "123"},
		{Kind: "invented", Label: "Claim", Value: "done"},
	} {
		if validateTaskEvidence(q) == nil {
			t.Fatalf("accepted invalid evidence: %s", q.Kind)
		}
	}
}
func TestTaskEvidenceAccessEnforcesScopeOwnershipAndTaskMembership(t *testing.T) {
	owner, appID, taskID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeApplicationRepo{app: &Application{Base: Base{ID: appID}, UserID: owner}, tasks: []ApplicationTask{{Base: Base{ID: taskID}, ApplicationID: appID, Status: "todo"}}}
	service := &Service{repo: repo}
	for _, tc := range []struct {
		name    string
		actor   principal.Principal
		task    uuid.UUID
		allowed bool
	}{
		{"owner", principal.Principal{Type: principal.PrincipalUser, UserID: &owner}, taskID, true},
		{"foreign owner", principal.Principal{Type: principal.PrincipalUser, UserID: func() *uuid.UUID { v := uuid.New(); return &v }()}, taskID, false},
		{"agent without tasks scope", principal.Principal{Type: principal.PrincipalAgent, UserID: &owner}, taskID, false},
		{"scoped agent", principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeTasks: {}}}, taskID, true},
		{"task from another application", principal.Principal{Type: principal.PrincipalUser, UserID: &owner}, uuid.New(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.taskEvidenceAccess(principal.WithPrincipal(context.Background(), tc.actor), appID, tc.task)
			if (err == nil) != tc.allowed {
				t.Fatalf("allowed=%v err=%v", tc.allowed, err)
			}
		})
	}
	if _, err := service.taskEvidenceAccess(context.Background(), appID, taskID); err == nil {
		t.Fatal("unauthenticated access allowed")
	}
}
