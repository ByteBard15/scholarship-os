package auth

import (
	"context"

	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	"github.com/example/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
)

type Repository interface {
	CreateAudit(context.Context, *profile.AuditLog) error
	CreateUser(context.Context, *user.User, *profile.ApplicantProfile, *profile.AuditLog) error
	GetUserByEmail(context.Context, string) (*user.User, error)
	GetUserByID(context.Context, uuid.UUID) (*user.User, error)
	ListUsers(context.Context) ([]user.User, error)
	UpdateUser(context.Context, *user.User, *profile.AuditLog) error
	CreateSession(context.Context, *UserSession, *user.User, *profile.AuditLog) error
	GetSessionByHash(context.Context, string) (*UserSession, error)
	TouchSession(context.Context, uuid.UUID) error
	RevokeSession(context.Context, uuid.UUID, *profile.AuditLog) error
	RevokeUserSessions(context.Context, uuid.UUID, *uuid.UUID, *profile.AuditLog) error
	UpdatePasswordAndRevokeSessions(context.Context, *user.User, *uuid.UUID, *profile.AuditLog) error
	CreateAgentCredential(context.Context, *AgentCredential, []AgentCredentialScope, *profile.AuditLog) error
	GetAgentCredentialByHash(context.Context, string) (*AgentCredential, error)
	GetAgentCredential(context.Context, uuid.UUID) (*AgentCredential, error)
	ListAgentCredentials(context.Context) ([]AgentCredential, error)
	TouchAgentCredential(context.Context, uuid.UUID) error
	RevokeAgentCredential(context.Context, *AgentCredential, *profile.AuditLog) error
}
