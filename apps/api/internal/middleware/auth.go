package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	featureauth "github.com/example/scholarship-os/apps/api/internal/features/auth"
	principal "github.com/example/scholarship-os/apps/api/pkg/auth"
	"github.com/example/scholarship-os/apps/api/pkg/response"
)

type TokenAuthenticator interface {
	Authenticate(context.Context, string) (principal.Principal, error)
}

func Authenticate(authenticator TokenAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required")
				return
			}
			actor, err := authenticator.Authenticate(r.Context(), parts[1])
			if err != nil {
				writeAuthenticationError(w, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(principal.WithPrincipal(r.Context(), actor)))
		})
	}
}

func RequireUserOrSystem(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := principal.PrincipalFromContext(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required")
			return
		}
		if actor.Type != principal.PrincipalUser && !actor.IsSystem() {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "user or system authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireSystem(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := principal.PrincipalFromContext(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required")
			return
		}
		if !actor.IsSystem() {
			response.Error(w, http.StatusForbidden, "SYSTEM_AUTH_REQUIRED", "system authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAgentScope(scope string) func(http.Handler) http.Handler {
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
			if actor.Type != principal.PrincipalAgent {
				response.Error(w, http.StatusForbidden, "AGENT_AUTH_REQUIRED", "agent authentication required")
				return
			}
			if !actor.HasScope(scope) {
				response.Error(w, http.StatusForbidden, "AGENT_SCOPE_REQUIRED", "required agent scope is missing")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthenticationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, featureauth.ErrTokenExpired):
		response.Error(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "token expired")
	case errors.Is(err, featureauth.ErrSessionRevoked):
		response.Error(w, http.StatusUnauthorized, "SESSION_REVOKED", "session revoked")
	case errors.Is(err, featureauth.ErrUserInactive):
		response.Error(w, http.StatusUnauthorized, "USER_INACTIVE", "user is inactive")
	case errors.Is(err, featureauth.ErrAgentKeyRevoked):
		response.Error(w, http.StatusUnauthorized, "AGENT_KEY_REVOKED", "agent key revoked")
	default:
		response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid authentication token")
	}
}
