package ownership

import (
	"context"
	"log/slog"
	"net/http"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OwnerLookup func(context.Context, uuid.UUID) (uuid.UUID, error)

// Middleware exposes resource-specific authorization factories. Each route
// names the operation it protects, so the ownership boundary is visible beside
// the handler instead of being inferred from the URL at runtime.
type Middleware struct {
	repository Repository
}

func NewMiddleware(repository Repository) *Middleware {
	return &Middleware{repository: repository}
}

func (m *Middleware) User(action string) func(http.Handler) http.Handler {
	return m.require(action, "user", "userID", func(_ context.Context, id uuid.UUID) (uuid.UUID, error) { return id, nil })
}

func (m *Middleware) Profile(action string) func(http.Handler) http.Handler {
	return m.ProfileParam(action, "profileID")
}

func (m *Middleware) ProfileParam(action, parameter string) func(http.Handler) http.Handler {
	return m.require(action, "profile", parameter, m.repository.ProfileOwner)
}

func (m *Middleware) Application(action string) func(http.Handler) http.Handler {
	return m.require(action, "application", "applicationID", m.repository.ApplicationOwner)
}

func (m *Middleware) ApplicationProposal(action string) func(http.Handler) http.Handler {
	return m.require(action, "application proposal", "proposalID", m.repository.ApplicationProposalOwner)
}

func (m *Middleware) Questionnaire(action string) func(http.Handler) http.Handler {
	return m.require(action, "questionnaire", "questionnaireID", m.repository.QuestionnaireOwner)
}

func (m *Middleware) Question(action string) func(http.Handler) http.Handler {
	return m.require(action, "question", "questionID", m.repository.QuestionOwner)
}

func (m *Middleware) InformationRequest(action string) func(http.Handler) http.Handler {
	return m.require(action, "information request", "requestID", m.repository.InformationRequestOwner)
}

func (m *Middleware) ResearchTask(action string) func(http.Handler) http.Handler {
	return m.require(action, "research task", "taskID", m.repository.ResearchTaskOwner)
}

func (m *Middleware) require(action, resource, parameter string, lookup OwnerLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := principal.PrincipalFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required")
				return
			}
			if actor.IsSystem() {
				next.ServeHTTP(w, r)
				return
			}
			if actor.UserID == nil {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "credential is not associated with a user workspace")
				return
			}

			id, err := uuid.Parse(chi.URLParam(r, parameter))
			if err != nil {
				response.Error(w, http.StatusBadRequest, "INVALID_ID", parameter+" must be a UUID")
				return
			}
			ownerID, err := lookup(r.Context(), id)
			if err != nil {
				appLogger.Warn(r.Context(), "ownership lookup failed", slog.String("action", action), slog.String("resource", resource), slog.String("resourceId", id.String()), slog.Any("error", err))
				response.Error(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "resource not found")
				return
			}
			if ownerID != *actor.UserID {
				appLogger.Warn(r.Context(), "ownership check denied", slog.String("action", action), slog.String("resource", resource), slog.String("resourceId", id.String()))
				response.Error(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "resource not found")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
