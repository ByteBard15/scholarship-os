package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/google/uuid"
)

type UpdateAgentScopesRequest struct {
	Scopes []string `json:"scopes" validate:"required,min=1,dive,required"`
}

func (s *Service) UpdateAgentScopes(ctx context.Context, id uuid.UUID, req UpdateAgentScopesRequest) (*AgentCredentialResponse, error) {
	if _, err := principal.RequireSystem(ctx); err != nil {
		return nil, err
	}
	scopes, err := normalizedScopes(req.Scopes)
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return nil, ErrInvalidAgentScope
	}
	value, err := s.repo.GetAgentCredential(ctx, id)
	if err != nil {
		return nil, err
	}
	if !value.IsActive || value.RevokedAt != nil {
		return nil, ErrAgentKeyRevoked
	}
	rows := make([]AgentCredentialScope, 0, len(scopes))
	for _, scope := range scopes {
		rows = append(rows, AgentCredentialScope{ID: uuid.New(), AgentCredentialID: id, Scope: scope})
	}
	metadata, _ := json.Marshal(map[string]any{"scopes": scopes})
	audit := &profile.AuditLog{UserID: value.UserID, Action: "AGENT_SCOPES_UPDATED", EntityType: strPointer("agent_credential"), EntityID: &id, Metadata: profile.JSON(metadata)}
	if err = s.repo.UpdateAgentScopes(ctx, id, rows, audit); err != nil {
		return nil, err
	}
	value.Scopes = rows
	out := agentCredentialResponse(value)
	return &out, nil
}
func strPointer(s string) *string { return &s }
func (h *Handler) UpdateAgentScopes(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "credentialID")
	if !ok {
		return
	}
	var req UpdateAgentScopesRequest
	if !h.decode(w, r, &req) {
		return
	}
	out, err := h.service.UpdateAgentScopes(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, out)
}

type AgentCapabilitiesResponse struct {
	AgentID           *uuid.UUID `json:"agentId,omitempty"`
	UserID            *uuid.UUID `json:"userId,omitempty"`
	Scopes            []string   `json:"scopes"`
	PreparationScopes []string   `json:"preparationScopes"`
	MissingScopes     []string   `json:"missingScopes"`
	PreparationReady  bool       `json:"preparationReady"`
	HumanSteps        []string   `json:"humanSteps"`
}

func AgentCapabilities(ctx context.Context) (*AgentCapabilitiesResponse, error) {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if actor.Type != principal.PrincipalAgent {
		return nil, principal.ErrForbidden
	}
	required := []string{principal.ScopeApplications, principal.ScopeInformationRequests, principal.ScopeProfiles, principal.ScopeQuestionnaires, principal.ScopeResearch, principal.ScopeTasks, principal.ScopeWriting}
	scopes := make([]string, 0, len(actor.Scopes))
	for scope := range actor.Scopes {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	missing := []string{}
	for _, scope := range required {
		if !actor.HasScope(scope) {
			missing = append(missing, scope)
		}
	}
	return &AgentCapabilitiesResponse{AgentID: actor.AgentID, UserID: actor.UserID, Scopes: scopes, PreparationScopes: required, MissingScopes: missing, PreparationReady: len(missing) == 0 && actor.UserID != nil, HumanSteps: []string{"proposal_approval", "writing_approval", "referee_sign_off", "external_messages", "application_submission"}}, nil
}
func (h *Handler) AgentCapabilities(w http.ResponseWriter, r *http.Request) {
	out, err := AgentCapabilities(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, out)
}
