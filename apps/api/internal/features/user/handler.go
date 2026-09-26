package user

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		response.Error(w, 400, "INVALID_USER_ID", "user ID must be a UUID")
		return
	}
	u, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, ToResponse(u))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, 400, "INVALID_JSON", "request body must be valid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, 400, "VALIDATION_ERROR", err.Error())
		return
	}
	u, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, ToResponse(u))
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, 404, "USER_NOT_FOUND", ErrNotFound.Error())
	case errors.Is(err, ErrEmailConflict):
		response.Error(w, 409, "EMAIL_CONFLICT", ErrEmailConflict.Error())
	default:
		appLogger.Error(r.Context(), "user request failed", "error", err)
		response.Error(w, 500, "INTERNAL_ERROR", "internal server error")
	}
}
