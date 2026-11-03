package auth

import (
	"context"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type accessRepository struct {
	Repository
	credential *AgentCredential
	scopes     []AgentCredentialScope
	audit      *profile.AuditLog
	writes     int
}

func (r *accessRepository) GetAgentCredential(context.Context, uuid.UUID) (*AgentCredential, error) {
	return r.credential, nil
}
func (r *accessRepository) UpdateAgentScopes(_ context.Context, _ uuid.UUID, scopes []AgentCredentialScope, audit *profile.AuditLog) error {
	r.scopes = scopes
	r.audit = audit
	r.writes++
	return nil
}
func TestOnlySystemCanAmendCredentialScopes(t *testing.T) {
	owner, id := uuid.New(), uuid.New()
	repo := &accessRepository{credential: &AgentCredential{ID: id, UserID: &owner, IsActive: true, KeyHash: "unchanged-hash"}}
	service := &Service{repo: repo}
	req := UpdateAgentScopesRequest{Scopes: []string{"writing", "research", "writing"}}
	for _, kind := range []principal.PrincipalType{principal.PrincipalUser, principal.PrincipalAgent} {
		ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: kind, UserID: &owner})
		if _, err := service.UpdateAgentScopes(ctx, id, req); err != principal.ErrSystemRequired || repo.writes != 0 {
			t.Fatal("self escalation allowed", err)
		}
	}
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalSystem})
	result, err := service.UpdateAgentScopes(ctx, id, req)
	if err != nil || len(result.Scopes) != 2 || repo.writes != 1 || repo.audit.Action != "AGENT_SCOPES_UPDATED" || repo.credential.KeyHash != "unchanged-hash" {
		t.Fatalf("scope update: %+v %v", result, err)
	}
	for _, scopes := range [][]string{{"root"}, {}} {
		if _, err := service.UpdateAgentScopes(ctx, id, UpdateAgentScopesRequest{Scopes: scopes}); err != ErrInvalidAgentScope {
			t.Fatal("invalid scope accepted", err)
		}
	}
	repo.credential.IsActive = false
	if _, err := service.UpdateAgentScopes(ctx, id, req); err != ErrAgentKeyRevoked {
		t.Fatal("revoked credential changed", err)
	}
}
func TestAgentCapabilitiesExplainMissingPermissions(t *testing.T) {
	owner := uuid.New()
	actor := principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeResearch: {}}}
	ctx := principal.WithPrincipal(context.Background(), actor)
	result, err := AgentCapabilities(ctx)
	if err != nil || result.PreparationReady || len(result.MissingScopes) != 6 || len(result.HumanSteps) != 5 {
		t.Fatalf("capabilities: %+v %v", result, err)
	}
	for _, scope := range result.PreparationScopes {
		actor.Scopes[scope] = struct{}{}
	}
	result, err = AgentCapabilities(principal.WithPrincipal(context.Background(), actor))
	if err != nil || !result.PreparationReady || len(result.MissingScopes) != 0 {
		t.Fatal("complete capability check", err)
	}
	if _, err := AgentCapabilities(context.Background()); err != principal.ErrAuthRequired {
		t.Fatal(err)
	}
	if _, err := AgentCapabilities(principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalUser})); err != principal.ErrForbidden {
		t.Fatal(err)
	}
}
