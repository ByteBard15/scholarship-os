package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service, validate *validator.Validate) *Handler {
	return &Handler{service: service, validate: validate}
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return false
	}
	if err := h.validate.Struct(value); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return false
	}
	return true
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.Register(r.Context(), request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, userResponse(value))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.Login(r.Context(), request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, value)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Logout(r.Context()); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.CurrentUser(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, userResponse(value))
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var request ChangePasswordRequest
	if !h.decode(w, r, &request) {
		return
	}
	if err := h.service.ChangePassword(r.Context(), request); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateAgentCredential(w http.ResponseWriter, r *http.Request) {
	var request CreateAgentCredentialRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.CreateAgentCredential(r.Context(), request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, value)
}

func (h *Handler) ListAgentCredentials(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListAgentCredentials(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, values, len(values))
}

func (h *Handler) RevokeAgentCredential(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "credentialID")
	if !ok {
		return
	}
	if err := h.service.RevokeAgentCredential(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListUsers(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, values, len(values))
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "userID")
	if !ok {
		return
	}
	value, err := h.service.GetAdminUser(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, value)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "userID")
	if !ok {
		return
	}
	var request UpdateUserRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.UpdateAdminUser(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, value)
}

func (h *Handler) SetUserActive(active bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "userID")
		if !ok {
			return
		}
		if err := h.service.SetUserActive(r.Context(), id, active); err != nil {
			h.writeError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "userID")
	if !ok {
		return
	}
	var request ResetPasswordRequest
	if !h.decode(w, r, &request) {
		return
	}
	if err := h.service.ResetPassword(r.Context(), id, request); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", name+" must be a UUID")
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
	case errors.Is(err, ErrInvalidToken):
		response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid authentication token")
	case errors.Is(err, ErrTokenExpired):
		response.Error(w, http.StatusUnauthorized, "TOKEN_EXPIRED", err.Error())
	case errors.Is(err, ErrSessionRevoked):
		response.Error(w, http.StatusUnauthorized, "SESSION_REVOKED", err.Error())
	case errors.Is(err, ErrUserInactive):
		response.Error(w, http.StatusUnauthorized, "USER_INACTIVE", err.Error())
	case errors.Is(err, ErrAgentKeyRevoked):
		response.Error(w, http.StatusUnauthorized, "AGENT_KEY_REVOKED", err.Error())
	case errors.Is(err, ErrEmailAlreadyExists):
		response.Error(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", err.Error())
	case errors.Is(err, ErrInvalidPassword):
		response.Error(w, http.StatusBadRequest, "INVALID_PASSWORD", err.Error())
	case errors.Is(err, ErrInvalidAgentScope):
		response.Error(w, http.StatusBadRequest, "INVALID_AGENT_SCOPE", err.Error())
	case errors.Is(err, ErrCredentialNotFound), errors.Is(err, user.ErrNotFound):
		response.Error(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "resource not found")
	case errors.Is(err, principal.ErrAuthRequired):
		response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", err.Error())
	case errors.Is(err, principal.ErrSystemRequired):
		response.Error(w, http.StatusForbidden, "SYSTEM_AUTH_REQUIRED", err.Error())
	case errors.Is(err, principal.ErrForbidden), errors.Is(err, ErrCurrentSessionRequired):
		response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		appLogger.Error(r.Context(), "authentication request failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
