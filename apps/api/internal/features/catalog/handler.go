package catalog

import (
	"encoding/json"
	"errors"
	"net/http"

	appLogger "github.com/example/scholarship-os/apps/api/pkg/logger"
	"github.com/example/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(s *Service, v *validator.Validate) *Handler { return &Handler{s, v} }
func parseID(r *http.Request, key, code string) (uuid.UUID, bool) {
	id, e := uuid.Parse(chi.URLParam(r, key))
	if e != nil {
		return uuid.Nil, false
	}
	return id, true
}
func decode(w http.ResponseWriter, r *http.Request, d any) bool {
	if e := json.NewDecoder(r.Body).Decode(d); e != nil {
		response.Error(w, 400, "INVALID_JSON", "request body must be valid JSON")
		return false
	}
	return true
}
func filters(r *http.Request) Filters {
	var f Filters
	if value := r.URL.Query().Get("institutionId"); value != "" {
		if id, e := uuid.Parse(value); e == nil {
			f.InstitutionID = &id
		}
	}
	if value := r.URL.Query().Get("country"); value != "" {
		f.Country = &value
	}
	if value := r.URL.Query().Get("degreeLevel"); value != "" {
		f.DegreeLevel = &value
	}
	return f
}
func (h *Handler) ListInstitutions(w http.ResponseWriter, r *http.Request) {
	v, e := h.service.ListInstitutions(r.Context())
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]InstitutionResponse, len(v))
	for i := range v {
		out[i] = toInstitution(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateInstitution(w http.ResponseWriter, r *http.Request) {
	var q InstitutionRequest
	if !decode(w, r, &q) {
		return
	}
	if e := h.validate.Struct(q); e != nil {
		response.Error(w, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	v, e := h.service.CreateInstitution(r.Context(), q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toInstitution(v))
}
func (h *Handler) GetInstitution(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "institutionID", "")
	if !ok {
		response.Error(w, 400, "INVALID_INSTITUTION_ID", "institution ID must be a UUID")
		return
	}
	v, e := h.service.GetInstitution(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toInstitution(v))
}
func (h *Handler) UpdateInstitution(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "institutionID", "")
	if !ok {
		response.Error(w, 400, "INVALID_INSTITUTION_ID", "institution ID must be a UUID")
		return
	}
	var q UpdateInstitutionRequest
	if !decode(w, r, &q) {
		return
	}
	if e := h.validate.Struct(q); e != nil {
		response.Error(w, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	v, e := h.service.UpdateInstitution(r.Context(), id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toInstitution(v))
}
func (h *Handler) ListProgrammes(w http.ResponseWriter, r *http.Request) {
	v, e := h.service.ListProgrammes(r.Context(), filters(r))
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ProgrammeResponse, len(v))
	for i := range v {
		out[i] = toProgramme(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateProgramme(w http.ResponseWriter, r *http.Request) {
	var q ProgrammeRequest
	if !decode(w, r, &q) {
		return
	}
	if e := h.validate.Struct(q); e != nil {
		response.Error(w, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	v, e := h.service.CreateProgramme(r.Context(), q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toProgramme(v))
}
func (h *Handler) GetProgramme(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "programmeID", "")
	if !ok {
		response.Error(w, 400, "INVALID_PROGRAMME_ID", "programme ID must be a UUID")
		return
	}
	v, e := h.service.GetProgramme(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toProgramme(v))
}
func (h *Handler) UpdateProgramme(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "programmeID", "")
	if !ok {
		response.Error(w, 400, "INVALID_PROGRAMME_ID", "programme ID must be a UUID")
		return
	}
	var q UpdateProgrammeRequest
	if !decode(w, r, &q) {
		return
	}
	if e := h.validate.Struct(q); e != nil {
		response.Error(w, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	v, e := h.service.UpdateProgramme(r.Context(), id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toProgramme(v))
}
func (h *Handler) ListScholarships(w http.ResponseWriter, r *http.Request) {
	v, e := h.service.ListScholarships(r.Context(), filters(r))
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ScholarshipResponse, len(v))
	for i := range v {
		out[i] = toScholarship(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateScholarship(w http.ResponseWriter, r *http.Request) {
	var q ScholarshipRequest
	if !decode(w, r, &q) {
		return
	}
	if e := h.validate.Struct(q); e != nil {
		response.Error(w, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	v, e := h.service.CreateScholarship(r.Context(), q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toScholarship(v))
}
func (h *Handler) GetScholarship(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "scholarshipID", "")
	if !ok {
		response.Error(w, 400, "INVALID_SCHOLARSHIP_ID", "scholarship ID must be a UUID")
		return
	}
	v, e := h.service.GetScholarship(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toScholarship(v))
}
func (h *Handler) UpdateScholarship(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "scholarshipID", "")
	if !ok {
		response.Error(w, 400, "INVALID_SCHOLARSHIP_ID", "scholarship ID must be a UUID")
		return
	}
	var q UpdateScholarshipRequest
	if !decode(w, r, &q) {
		return
	}
	if e := h.validate.Struct(q); e != nil {
		response.Error(w, 400, "VALIDATION_ERROR", e.Error())
		return
	}
	v, e := h.service.UpdateScholarship(r.Context(), id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toScholarship(v))
}
func (h *Handler) err(w http.ResponseWriter, r *http.Request, e error) {
	switch {
	case errors.Is(e, ErrInstitutionNotFound):
		response.Error(w, 404, "INSTITUTION_NOT_FOUND", e.Error())
	case errors.Is(e, ErrProgrammeNotFound):
		response.Error(w, 404, "PROGRAMME_NOT_FOUND", e.Error())
	case errors.Is(e, ErrScholarshipNotFound):
		response.Error(w, 404, "SCHOLARSHIP_NOT_FOUND", e.Error())
	default:
		appLogger.Error(r.Context(), "catalog request failed", "error", e)
		response.Error(w, 500, "INTERNAL_ERROR", "internal server error")
	}
}
