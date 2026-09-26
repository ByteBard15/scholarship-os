package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

var allowedAgentScopes = map[string]struct{}{
	principal.ScopeResearch: {}, principal.ScopeApplications: {}, principal.ScopeQuestionnaires: {},
	principal.ScopeInformationRequests: {}, principal.ScopeProfiles: {}, principal.ScopeTasks: {},
}

type Service struct {
	repo        Repository
	hasher      PasswordHasher
	systemKey   string
	sessionTTL  time.Duration
	tokenPepper string
	dummyHash   string
	now         func() time.Time
}

func NewService(repo Repository, hasher PasswordHasher, systemKey string, sessionTTL time.Duration, tokenPepper string) *Service {
	dummy, _ := hasher.Hash("invalid-dummy-password")
	return &Service{repo: repo, hasher: hasher, systemKey: systemKey, sessionTTL: sessionTTL, tokenPepper: tokenPepper, dummyHash: dummy, now: func() time.Time { return time.Now().UTC() }}
}

func NormalizeEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func (s *Service) Register(ctx context.Context, request RegisterRequest) (*user.User, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	if len(request.Password) < 10 {
		return nil, ErrInvalidPassword
	}
	email := NormalizeEmail(request.Email)
	if existing, err := s.repo.GetUserByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, user.ErrNotFound) {
		return nil, err
	}
	hash, err := s.hasher.Hash(request.Password)
	if err != nil {
		return nil, err
	}
	value := &user.User{ID: uuid.New(), Email: email, DisplayName: request.DisplayName, PasswordHash: hash, IsActive: true}
	createProfile := request.CreateDefaultProfile == nil || *request.CreateDefaultProfile
	var master *profile.ApplicantProfile
	if createProfile {
		master = &profile.ApplicantProfile{Base: profile.Base{ID: uuid.New()}, UserID: value.ID, Name: "Master Profile", ProfileType: profile.ProfileTypeMaster, IsDefault: true}
	}
	audit := authAudit(&value.ID, "USER_REGISTERED", "user", &value.ID)
	if err := s.repo.CreateUser(ctx, value, master, audit); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (*LoginResponse, error) {
	email := NormalizeEmail(request.Email)
	value, err := s.repo.GetUserByEmail(ctx, email)
	encoded := s.dummyHash
	if err == nil && value.PasswordHash != "" {
		encoded = value.PasswordHash
	}
	valid, verifyErr := s.hasher.Verify(encoded, request.Password)
	if err != nil || verifyErr != nil || !valid || value == nil || !value.IsActive || value.PasswordHash == "" {
		_ = s.repo.CreateAudit(ctx, authAudit(nil, "LOGIN_FAILURE", "authentication", nil))
		return nil, ErrInvalidCredentials
	}
	token, err := secureToken("usr_")
	if err != nil {
		return nil, err
	}
	now := s.now()
	value.LastLoginAt = &now
	session := &UserSession{ID: uuid.New(), UserID: value.ID, TokenHash: s.hashToken(token), ExpiresAt: now.Add(s.sessionTTL), CreatedAt: now}
	audit := authAudit(&value.ID, "LOGIN_SUCCESS", "user_session", &session.ID)
	if err := s.repo.CreateSession(ctx, session, value, audit); err != nil {
		return nil, err
	}
	return &LoginResponse{AccessToken: token, TokenType: "Bearer", ExpiresAt: session.ExpiresAt, User: userResponse(value)}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (principal.Principal, error) {
	switch {
	case strings.HasPrefix(token, "sys_"):
		if s.systemKey == "" || !constantTimeEqual(token, s.systemKey) {
			return principal.Principal{}, ErrInvalidToken
		}
		return principal.Principal{Type: principal.PrincipalSystem, Scopes: map[string]struct{}{}}, nil
	case strings.HasPrefix(token, "usr_"):
		return s.authenticateSession(ctx, token)
	case strings.HasPrefix(token, "agt_"):
		return s.authenticateAgent(ctx, token)
	default:
		return principal.Principal{}, ErrInvalidToken
	}
}

func (s *Service) authenticateSession(ctx context.Context, token string) (principal.Principal, error) {
	session, err := s.repo.GetSessionByHash(ctx, s.hashToken(token))
	if err != nil {
		return principal.Principal{}, ErrInvalidToken
	}
	if session.RevokedAt != nil {
		return principal.Principal{}, ErrSessionRevoked
	}
	if !session.ExpiresAt.After(s.now()) {
		return principal.Principal{}, ErrTokenExpired
	}
	if !session.User.IsActive {
		return principal.Principal{}, ErrUserInactive
	}
	_ = s.repo.TouchSession(ctx, session.ID)
	return principal.Principal{Type: principal.PrincipalUser, UserID: &session.UserID, SessionID: &session.ID, Scopes: map[string]struct{}{}}, nil
}

func (s *Service) authenticateAgent(ctx context.Context, token string) (principal.Principal, error) {
	credential, err := s.repo.GetAgentCredentialByHash(ctx, s.hashToken(token))
	if err != nil {
		return principal.Principal{}, ErrInvalidToken
	}
	if !credential.IsActive || credential.RevokedAt != nil {
		return principal.Principal{}, ErrAgentKeyRevoked
	}
	if credential.ExpiresAt != nil && !credential.ExpiresAt.After(s.now()) {
		return principal.Principal{}, ErrTokenExpired
	}
	if credential.UserID != nil && (credential.User == nil || !credential.User.IsActive) {
		return principal.Principal{}, ErrUserInactive
	}
	scopes := make(map[string]struct{}, len(credential.Scopes))
	for _, scope := range credential.Scopes {
		scopes[scope.Scope] = struct{}{}
	}
	_ = s.repo.TouchAgentCredential(ctx, credential.ID)
	return principal.Principal{Type: principal.PrincipalAgent, UserID: credential.UserID, AgentID: &credential.ID, Scopes: scopes}, nil
}

func (s *Service) CurrentUser(ctx context.Context) (*user.User, error) {
	actor, err := principal.RequireUser(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(ctx, *actor.UserID)
}

func (s *Service) Logout(ctx context.Context) error {
	actor, err := principal.RequireUser(ctx)
	if err != nil || actor.SessionID == nil {
		return ErrCurrentSessionRequired
	}
	return s.repo.RevokeSession(ctx, *actor.SessionID, authAudit(actor.UserID, "LOGOUT", "user_session", actor.SessionID))
}

func (s *Service) ChangePassword(ctx context.Context, request ChangePasswordRequest) error {
	actor, err := principal.RequireUser(ctx)
	if err != nil || actor.SessionID == nil {
		return ErrCurrentSessionRequired
	}
	if len(request.NewPassword) < 10 {
		return ErrInvalidPassword
	}
	value, err := s.repo.GetUserByID(ctx, *actor.UserID)
	if err != nil {
		return err
	}
	valid, verifyErr := s.hasher.Verify(value.PasswordHash, request.CurrentPassword)
	if verifyErr != nil || !valid {
		return ErrInvalidCredentials
	}
	value.PasswordHash, err = s.hasher.Hash(request.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePasswordAndRevokeSessions(ctx, value, actor.SessionID, authAudit(&value.ID, "PASSWORD_CHANGED", "user", &value.ID))
}

func (s *Service) CreateAgentCredential(ctx context.Context, request CreateAgentCredentialRequest) (*CreatedAgentCredentialResponse, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	if request.UserID == nil {
		return nil, errors.New("userId is required for Phase 4.5 agent credentials")
	}
	owner, err := s.repo.GetUserByID(ctx, *request.UserID)
	if err != nil {
		return nil, err
	}
	if !owner.IsActive {
		return nil, ErrUserInactive
	}
	scopes, err := normalizedScopes(request.Scopes)
	if err != nil {
		return nil, err
	}
	token, err := secureToken("agt_")
	if err != nil {
		return nil, err
	}
	prefixLength := 12
	if len(token) < prefixLength {
		prefixLength = len(token)
	}
	credential := &AgentCredential{ID: uuid.New(), UserID: request.UserID, Name: strings.TrimSpace(request.Name), KeyPrefix: token[:prefixLength], KeyHash: s.hashToken(token), IsActive: true, ExpiresAt: request.ExpiresAt}
	rows := make([]AgentCredentialScope, 0, len(scopes))
	for _, scope := range scopes {
		rows = append(rows, AgentCredentialScope{ID: uuid.New(), AgentCredentialID: credential.ID, Scope: scope})
	}
	entityType := "agent_credential"
	audit := authAudit(request.UserID, "AGENT_CREDENTIAL_CREATED", entityType, &credential.ID)
	if err := s.repo.CreateAgentCredential(ctx, credential, rows, audit); err != nil {
		return nil, err
	}
	credential.Scopes = rows
	response := CreatedAgentCredentialResponse{AgentCredentialResponse: agentCredentialResponse(credential), APIKey: token}
	return &response, nil
}

func (s *Service) ListAgentCredentials(ctx context.Context) ([]AgentCredentialResponse, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	values, err := s.repo.ListAgentCredentials(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]AgentCredentialResponse, 0, len(values))
	for i := range values {
		result = append(result, agentCredentialResponse(&values[i]))
	}
	return result, nil
}

func (s *Service) RevokeAgentCredential(ctx context.Context, id uuid.UUID) error {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return err
	}
	value, err := s.repo.GetAgentCredential(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.RevokeAgentCredential(ctx, value, authAudit(value.UserID, "AGENT_CREDENTIAL_REVOKED", "agent_credential", &value.ID))
}

func (s *Service) ListUsers(ctx context.Context) ([]UserResponse, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	values, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]UserResponse, len(values))
	for i := range values {
		result[i] = userResponse(&values[i])
	}
	return result, nil
}

func (s *Service) GetAdminUser(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	value, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := userResponse(value)
	return &response, nil
}

func (s *Service) UpdateAdminUser(ctx context.Context, id uuid.UUID, request UpdateUserRequest) (*UserResponse, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	value, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	value.DisplayName = request.DisplayName
	if err := s.repo.UpdateUser(ctx, value, nil); err != nil {
		return nil, err
	}
	response := userResponse(value)
	return &response, nil
}

func (s *Service) SetUserActive(ctx context.Context, id uuid.UUID, active bool) error {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return err
	}
	value, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	value.IsActive = active
	action := "USER_DISABLED"
	if active {
		action = "USER_ENABLED"
	}
	if err := s.repo.UpdateUser(ctx, value, authAudit(&value.ID, action, "user", &value.ID)); err != nil {
		return err
	}
	if !active {
		return s.repo.RevokeUserSessions(ctx, value.ID, nil, nil)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, id uuid.UUID, request ResetPasswordRequest) error {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return err
	}
	if len(request.NewPassword) < 10 {
		return ErrInvalidPassword
	}
	value, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	value.PasswordHash, err = s.hasher.Hash(request.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePasswordAndRevokeSessions(ctx, value, nil, authAudit(&value.ID, "PASSWORD_RESET_BY_SYSTEM", "user", &value.ID))
}

func (s *Service) hashToken(token string) string {
	sum := sha256.Sum256([]byte(token + s.tokenPepper))
	return hex.EncodeToString(sum[:])
}

func secureToken(prefix string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

func constantTimeEqual(left, right string) bool {
	a := sha256.Sum256([]byte(left))
	b := sha256.Sum256([]byte(right))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

func normalizedScopes(values []string) ([]string, error) {
	unique := map[string]struct{}{}
	for _, raw := range values {
		scope := strings.ToLower(strings.TrimSpace(raw))
		if _, ok := allowedAgentScopes[scope]; !ok {
			return nil, ErrInvalidAgentScope
		}
		unique[scope] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for scope := range unique {
		result = append(result, scope)
	}
	sort.Strings(result)
	return result, nil
}

func agentCredentialResponse(value *AgentCredential) AgentCredentialResponse {
	scopes := make([]string, 0, len(value.Scopes))
	for _, scope := range value.Scopes {
		scopes = append(scopes, scope.Scope)
	}
	sort.Strings(scopes)
	return AgentCredentialResponse{ID: value.ID, UserID: value.UserID, Name: value.Name, KeyPrefix: value.KeyPrefix, Scopes: scopes, IsActive: value.IsActive, ExpiresAt: value.ExpiresAt, LastUsedAt: value.LastUsedAt, RevokedAt: value.RevokedAt, CreatedAt: value.CreatedAt}
}

func authAudit(userID *uuid.UUID, action, entityType string, entityID *uuid.UUID) *profile.AuditLog {
	return &profile.AuditLog{ID: uuid.New(), UserID: userID, Action: action, EntityType: &entityType, EntityID: entityID}
}
