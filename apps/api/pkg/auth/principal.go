package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type PrincipalType string

const (
	PrincipalUser   PrincipalType = "user"
	PrincipalAgent  PrincipalType = "agent"
	PrincipalSystem PrincipalType = "system"
)

const (
	ScopeResearch            = "research"
	ScopeApplications        = "applications"
	ScopeQuestionnaires      = "questionnaires"
	ScopeInformationRequests = "information_requests"
	ScopeProfiles            = "profiles"
	ScopeTasks               = "tasks"
	ScopeWriting             = "writing"
)

var (
	ErrAuthRequired       = errors.New("authentication required")
	ErrForbidden          = errors.New("forbidden")
	ErrSystemRequired     = errors.New("system authentication required")
	ErrAgentRequired      = errors.New("agent authentication required")
	ErrAgentScopeRequired = errors.New("agent scope required")
)

type Principal struct {
	Type      PrincipalType
	UserID    *uuid.UUID
	AgentID   *uuid.UUID
	SessionID *uuid.UUID
	Scopes    map[string]struct{}
}

func (p Principal) IsSystem() bool { return p.Type == PrincipalSystem }
func (p Principal) HasScope(scope string) bool {
	if p.IsSystem() {
		return true
	}
	_, ok := p.Scopes[scope]
	return ok
}

type contextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(contextKey{}).(Principal)
	return principal, ok
}

func RequireAuthenticated(ctx context.Context) (Principal, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return Principal{}, ErrAuthRequired
	}
	return principal, nil
}

func RequireUser(ctx context.Context) (Principal, error) {
	principal, err := RequireAuthenticated(ctx)
	if err != nil {
		return Principal{}, err
	}
	if principal.Type != PrincipalUser || principal.UserID == nil {
		return Principal{}, ErrForbidden
	}
	return principal, nil
}

func RequireSystem(ctx context.Context) (Principal, error) {
	principal, err := RequireAuthenticated(ctx)
	if err != nil {
		return Principal{}, err
	}
	if !principal.IsSystem() {
		return Principal{}, ErrSystemRequired
	}
	return principal, nil
}

// EnforceOwner is used by domain services after loading user-owned resources.
// Calls without a request principal are treated as trusted internal calls.
func EnforceOwner(ctx context.Context, ownerID uuid.UUID) error {
	principal, ok := PrincipalFromContext(ctx)
	if !ok || principal.IsSystem() {
		return nil
	}
	if principal.UserID == nil || *principal.UserID != ownerID {
		return ErrForbidden
	}
	return nil
}
