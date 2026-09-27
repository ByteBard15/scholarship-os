package writing

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *Handler) ListSamples(w http.ResponseWriter, r *http.Request) {
	filters := SampleFilters{}
	if value := r.URL.Query().Get("documentType"); value != "" {
		filters.DocumentType = &value
	}
	if value := r.URL.Query().Get("applicationId"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_APPLICATION_ID", "applicationId must be a UUID")
			return
		}
		filters.ApplicationID = &id
	}
	items, err := h.service.ListSamples(r.Context(), queryUserID(r), filters)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, items, len(items))
}

func (h *Handler) CreateSample(w http.ResponseWriter, r *http.Request) {
	var request CreateSampleRequest
	if !h.decodeAndValidate(w, r, &request) {
		return
	}
	item, err := h.service.CreateSample(r.Context(), request, false)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, item)
}

func (h *Handler) GetSample(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "sampleID", "INVALID_WRITING_SAMPLE_ID")
	if !ok {
		return
	}
	item, err := h.service.GetSample(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, item)
}

func (h *Handler) UpdateSample(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "sampleID", "INVALID_WRITING_SAMPLE_ID")
	if !ok {
		return
	}
	var request UpdateSampleRequest
	if !h.decodeAndValidate(w, r, &request) {
		return
	}
	item, err := h.service.UpdateSample(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, item)
}

func (h *Handler) DeleteSample(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "sampleID", "INVALID_WRITING_SAMPLE_ID")
	if !ok {
		return
	}
	if err := h.service.DeleteSample(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListTags(r.Context(), queryUserID(r))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, items, len(items))
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var request CreateTagRequest
	if !h.decodeAndValidate(w, r, &request) {
		return
	}
	item, err := h.service.CreateTag(r.Context(), request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, item)
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "tagID", "INVALID_WRITING_TAG_ID")
	if !ok {
		return
	}
	if err := h.service.DeleteTag(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AgentListTags(w http.ResponseWriter, r *http.Request) {
	h.ListTags(w, r)
}

func (h *Handler) AgentMatchSamples(w http.ResponseWriter, r *http.Request) {
	var request MatchSamplesRequest
	if !h.decodeAndValidate(w, r, &request) {
		return
	}
	items, err := h.service.MatchSamples(r.Context(), request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, items, len(items))
}

func (h *Handler) AgentCreateGeneratedSample(w http.ResponseWriter, r *http.Request) {
	var request CreateSampleRequest
	if !h.decodeAndValidate(w, r, &request) {
		return
	}
	item, err := h.service.CreateSample(r.Context(), request, true)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, item)
}

func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return false
	}
	if err := h.validate.Struct(target); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return false
	}
	return true
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrSampleNotFound):
		response.Error(w, http.StatusNotFound, "WRITING_SAMPLE_NOT_FOUND", err.Error())
	case errors.Is(err, ErrTagNotFound):
		response.Error(w, http.StatusNotFound, "WRITING_TAG_NOT_FOUND", err.Error())
	case errors.Is(err, ErrUserRequired), errors.Is(err, ErrInvalidApplication), errors.Is(err, ErrInvalidSource), errors.Is(err, ErrTagsRequired):
		response.Error(w, http.StatusBadRequest, "INVALID_WRITING_REQUEST", err.Error())
	case errors.Is(err, principal.ErrForbidden):
		response.Error(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	default:
		appLogger.Error(r.Context(), "writing library request failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func pathID(w http.ResponseWriter, r *http.Request, key, code string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		response.Error(w, http.StatusBadRequest, code, key+" must be a UUID")
		return uuid.Nil, false
	}
	return id, true
}

func queryUserID(r *http.Request) *uuid.UUID {
	value := r.URL.Query().Get("userId")
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return &id
}
