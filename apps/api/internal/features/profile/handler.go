package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/byte/scholarship-os/apps/api/internal/features/user"
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
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := pathID(w, r, "userID")
	if !ok {
		return
	}
	var req CreateProfileRequest
	if !h.decode(w, r, &req) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, 400, "VALIDATION_ERROR", err.Error())
		return
	}
	p, err := h.service.Create(r.Context(), uid, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, profileResponse(p))
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := pathID(w, r, "userID")
	if !ok {
		return
	}
	rows, err := h.service.List(r.Context(), uid)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]ProfileResponse, 0, len(rows))
	for i := range rows {
		out = append(out, profileResponse(&rows[i]))
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	p, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, profileResponse(p))
}
func (h *Handler) GetFull(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	p, err := h.service.GetFull(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, fullResponse(p))
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req UpdateProfileRequest
	if !h.decode(w, r, &req) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, 400, "VALIDATION_ERROR", err.Error())
		return
	}
	p, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, profileResponse(p))
}
func (h *Handler) GetPersonalInfo(w http.ResponseWriter, r *http.Request) {
	pid, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	v, err := h.service.GetPersonalInfo(r.Context(), pid)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, personalResponse(v))
}
func (h *Handler) PutPersonalInfo(w http.ResponseWriter, r *http.Request) {
	h.writePersonal(w, r, false)
}
func (h *Handler) PatchPersonalInfo(w http.ResponseWriter, r *http.Request) {
	h.writePersonal(w, r, true)
}
func (h *Handler) writePersonal(w http.ResponseWriter, r *http.Request, patch bool) {
	pid, ok := pathID(w, r, "profileID")
	if !ok {
		return
	}
	var req PersonalInfoRequest
	if !h.decode(w, r, &req) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, 400, "VALIDATION_ERROR", err.Error())
		return
	}
	var v *ProfilePersonalInfo
	var err error
	if patch {
		v, err = h.service.PatchPersonalInfo(r.Context(), pid, req)
	} else {
		v, err = h.service.PutPersonalInfo(r.Context(), pid, req)
	}
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, personalResponse(v))
}

func (h *Handler) ListSection(k SectionKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := pathID(w, r, "profileID")
		if !ok {
			return
		}
		v, err := h.service.ListSection(r.Context(), k, pid)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		out := sectionResponses(v)
		response.Collection(w, out, len(out))
	}
}
func (h *Handler) GetSection(k SectionKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := pathID(w, r, "profileID")
		if !ok {
			return
		}
		id, ok := pathID(w, r, "sectionID")
		if !ok {
			return
		}
		v, err := h.service.GetSection(r.Context(), k, pid, id)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 200, sectionResponse(v))
	}
}
func (h *Handler) CreateSection(k SectionKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := pathID(w, r, "profileID")
		if !ok {
			return
		}
		var req SectionRequest
		if !h.decode(w, r, &req) {
			return
		}
		if err := h.validate.Struct(req); err != nil {
			response.Error(w, 400, "VALIDATION_ERROR", err.Error())
			return
		}
		v, err := h.service.CreateSection(r.Context(), k, pid, req)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 201, sectionResponse(v))
	}
}
func (h *Handler) UpdateSection(k SectionKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := pathID(w, r, "profileID")
		if !ok {
			return
		}
		id, ok := pathID(w, r, "sectionID")
		if !ok {
			return
		}
		var req SectionRequest
		if !h.decode(w, r, &req) {
			return
		}
		if err := h.validate.Struct(req); err != nil {
			response.Error(w, 400, "VALIDATION_ERROR", err.Error())
			return
		}
		v, err := h.service.UpdateSection(r.Context(), k, pid, id, req)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 200, sectionResponse(v))
	}
}
func (h *Handler) DeleteSection(k SectionKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := pathID(w, r, "profileID")
		if !ok {
			return
		}
		id, ok := pathID(w, r, "sectionID")
		if !ok {
			return
		}
		if err := h.service.DeleteSection(r.Context(), k, pid, id); err != nil {
			h.writeError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func (h *Handler) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		response.Error(w, 400, "INVALID_JSON", "request body must be valid JSON: "+err.Error())
		return false
	}
	return true
}
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, user.ErrNotFound):
		response.Error(w, 404, "USER_NOT_FOUND", user.ErrNotFound.Error())
	case errors.Is(err, ErrProfileNotFound):
		response.Error(w, 404, "PROFILE_NOT_FOUND", ErrProfileNotFound.Error())
	case errors.Is(err, ErrSectionNotFound):
		response.Error(w, 404, "SECTION_NOT_FOUND", ErrSectionNotFound.Error())
	case errors.Is(err, ErrPersonalInfoNotFound):
		response.Error(w, 404, "PERSONAL_INFO_NOT_FOUND", ErrPersonalInfoNotFound.Error())
	case errors.Is(err, ErrSkillConflict):
		response.Error(w, 409, "SKILL_CONFLICT", ErrSkillConflict.Error())
	case errors.Is(err, ErrProfileParentRequired):
		response.Error(w, http.StatusBadRequest, "PROFILE_PARENT_REQUIRED", err.Error())
	case errors.Is(err, ErrInvalidProfileParent):
		response.Error(w, http.StatusBadRequest, "INVALID_PROFILE_PARENT", err.Error())
	case errors.Is(err, ErrProfileInheritanceCycle):
		response.Error(w, http.StatusConflict, "PROFILE_INHERITANCE_CYCLE", err.Error())
	case errors.Is(err, ErrProfileOverrideNotFound):
		response.Error(w, http.StatusNotFound, "PROFILE_OVERRIDE_NOT_FOUND", err.Error())
	case errors.Is(err, ErrProfileDocumentNotFound):
		response.Error(w, http.StatusNotFound, "PROFILE_DOCUMENT_NOT_FOUND", err.Error())
	case errors.Is(err, ErrProfileImportNotFound):
		response.Error(w, http.StatusNotFound, "PROFILE_IMPORT_NOT_FOUND", err.Error())
	case errors.Is(err, ErrImportCandidateNotFound):
		response.Error(w, http.StatusNotFound, "IMPORT_CANDIDATE_NOT_FOUND", err.Error())
	case errors.Is(err, ErrImportNotReady):
		response.Error(w, http.StatusConflict, "IMPORT_NOT_READY", err.Error())
	case errors.Is(err, ErrInvalidDocumentType):
		response.Error(w, http.StatusBadRequest, "INVALID_DOCUMENT_TYPE", err.Error())
	case errors.Is(err, ErrInvalidDocument):
		response.Error(w, http.StatusBadRequest, "INVALID_DOCUMENT", err.Error())
	case errors.Is(err, ErrSnapshotNotFound):
		response.Error(w, http.StatusNotFound, "PROFILE_SNAPSHOT_NOT_FOUND", err.Error())
	case isValidationError(err):
		response.Error(w, 400, "VALIDATION_ERROR", err.Error())
	default:
		appLogger.Error(r.Context(), "profile request failed", "error", err)
		response.Error(w, 500, "INTERNAL_ERROR", "internal server error")
	}
}
func pathID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		response.Error(w, 400, "INVALID_ID", fmt.Sprintf("%s must be a UUID", key))
		return uuid.Nil, false
	}
	return id, true
}
