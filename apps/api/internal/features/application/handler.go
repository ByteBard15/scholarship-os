package application

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/example/scholarship-os/apps/api/internal/features/catalog"
	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	appLogger "github.com/example/scholarship-os/apps/api/pkg/logger"
	"github.com/example/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  *Service
	research *ResearchService
	validate *validator.Validate
}

func NewHandler(s *Service, r *ResearchService, v *validator.Validate) *Handler {
	return &Handler{s, r, v}
}
func (h *Handler) id(w http.ResponseWriter, r *http.Request, key, code string) (uuid.UUID, bool) {
	v, e := uuid.Parse(chi.URLParam(r, key))
	if e != nil {
		response.Error(w, 400, code, key+" must be a UUID")
		return uuid.Nil, false
	}
	return v, true
}
func (h *Handler) body(w http.ResponseWriter, r *http.Request, v any, validate bool) bool {
	if e := json.NewDecoder(r.Body).Decode(v); e != nil {
		response.Error(w, 400, "INVALID_JSON", "request body must be valid JSON")
		return false
	}
	if validate {
		if e := h.validate.Struct(v); e != nil {
			response.Error(w, 400, "VALIDATION_ERROR", e.Error())
			return false
		}
	}
	return true
}
func (h *Handler) appID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return h.id(w, r, "applicationID", "INVALID_APPLICATION_ID")
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var f Filters
	if v := r.URL.Query().Get("status"); v != "" {
		x := Status(v)
		f.Status = &x
	}
	if v := r.URL.Query().Get("country"); v != "" {
		f.Country = &v
	}
	if v := r.URL.Query().Get("intakeYear"); v != "" {
		if x, e := strconv.Atoi(v); e == nil {
			f.IntakeYear = &x
		} else {
			response.Error(w, 400, "VALIDATION_ERROR", "intakeYear must be an integer")
			return
		}
	}
	v, e := h.service.List(r.Context(), f)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ApplicationResponse, len(v))
	for i := range v {
		out[i] = toApplication(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var q CreateApplicationRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.Create(r.Context(), q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toApplication(v))
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.Detail(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, v)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q UpdateApplicationRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.Update(r.Context(), id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toApplication(v))
}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.appID(w, r)
	if !ok {
		return
	}
	if e := h.service.Delete(r.Context(), id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	id, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.Summary(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, v)
}
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	id, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.Readiness(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, v)
}
func (h *Handler) Eligibility(w http.ResponseWriter, r *http.Request) {
	id, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.Eligibility(r.Context(), id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, v)
}

func (h *Handler) ListRequirements(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListRequirements(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]RequirementResponse, len(v))
	for i := range v {
		out[i] = toRequirement(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateRequirement(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q RequirementRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateRequirement(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toRequirement(v))
}
func (h *Handler) GetRequirement(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "requirementID", "INVALID_REQUIREMENT_ID")
	if !ok {
		return
	}
	v, e := h.service.GetRequirement(r.Context(), a, id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toRequirement(v))
}
func (h *Handler) UpdateRequirement(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "requirementID", "INVALID_REQUIREMENT_ID")
	if !ok {
		return
	}
	var q RequirementRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.UpdateRequirement(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toRequirement(v))
}
func (h *Handler) DeleteRequirement(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "requirementID", "INVALID_REQUIREMENT_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteRequirement(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListDeadlines(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListDeadlines(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]DeadlineResponse, len(v))
	for i := range v {
		out[i] = toDeadline(&v[i], h.service.now())
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateDeadline(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q DeadlineRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateDeadline(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toDeadline(v, h.service.now()))
}
func (h *Handler) UpdateDeadline(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "deadlineID", "INVALID_DEADLINE_ID")
	if !ok {
		return
	}
	var q DeadlineRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.UpdateDeadline(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toDeadline(v, h.service.now()))
}
func (h *Handler) DeleteDeadline(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "deadlineID", "INVALID_DEADLINE_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteDeadline(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListFunding(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListFunding(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]FundingResponse, len(v))
	for i := range v {
		out[i] = toFunding(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateFunding(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q FundingRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateFunding(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toFunding(v))
}
func (h *Handler) UpdateFunding(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "fundingID", "INVALID_FUNDING_ID")
	if !ok {
		return
	}
	var q FundingRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.UpdateFunding(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toFunding(v))
}
func (h *Handler) DeleteFunding(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "fundingID", "INVALID_FUNDING_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteFunding(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListContacts(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ContactResponse, len(v))
	for i := range v {
		out[i] = toContact(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q ContactRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateContact(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toContact(v))
}
func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "contactID", "INVALID_CONTACT_ID")
	if !ok {
		return
	}
	var q ContactRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.UpdateContact(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toContact(v))
}
func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "contactID", "INVALID_CONTACT_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteContact(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListSupervisors(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListSupervisors(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]SupervisorResponse, len(v))
	for i := range v {
		out[i] = toSupervisor(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateSupervisor(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q SupervisorRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateSupervisor(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toSupervisor(v))
}
func (h *Handler) UpdateSupervisor(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "supervisorID", "INVALID_SUPERVISOR_ID")
	if !ok {
		return
	}
	var q SupervisorRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.UpdateSupervisor(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toSupervisor(v))
}
func (h *Handler) DeleteSupervisor(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "supervisorID", "INVALID_SUPERVISOR_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteSupervisor(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListURLs(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListURLs(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]URLResponse, len(v))
	for i := range v {
		out[i] = toURL(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q URLRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateURL(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toURL(v))
}
func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "urlID", "INVALID_URL_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteURL(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.service.ListTasks(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]TaskResponse, len(v))
	for i := range v {
		out[i] = toTask(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	var q TaskRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateTask(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toTask(v))
}
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "taskID", "INVALID_TASK_ID")
	if !ok {
		return
	}
	var q TaskRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.UpdateTask(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toTask(v))
}
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "taskID", "INVALID_TASK_ID")
	if !ok {
		return
	}
	if e := h.service.DeleteTask(r.Context(), a, id); e != nil {
		h.err(w, r, e)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) { h.setTask(w, r, true) }
func (h *Handler) ReopenTask(w http.ResponseWriter, r *http.Request)   { h.setTask(w, r, false) }
func (h *Handler) setTask(w http.ResponseWriter, r *http.Request, complete bool) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "taskID", "INVALID_TASK_ID")
	if !ok {
		return
	}
	v, e := h.service.SetTaskCompletion(r.Context(), a, id, complete)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toTask(v))
}

func (h *Handler) EvidenceSuggestions(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "requirementID", "INVALID_REQUIREMENT_ID")
	if !ok {
		return
	}
	v, e := h.service.EvidenceSuggestions(r.Context(), a, id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Collection(w, v, len(v))
}
func (h *Handler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "requirementID", "INVALID_REQUIREMENT_ID")
	if !ok {
		return
	}
	v, e := h.service.ListEvidence(r.Context(), a, id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]EvidenceResponse, len(v))
	for i, x := range v {
		out[i] = EvidenceResponse{x.ID, x.RequirementID, x.ProfileID, x.EntityType, x.EntityID, x.EvidenceID, x.Status, x.Notes, x.CreatedAt.UTC()}
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) CreateEvidence(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "requirementID", "INVALID_REQUIREMENT_ID")
	if !ok {
		return
	}
	var q EvidenceRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.service.CreateEvidence(r.Context(), a, id, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, EvidenceResponse{v.ID, v.RequirementID, v.ProfileID, v.EntityType, v.EntityID, v.EvidenceID, v.Status, v.Notes, v.CreatedAt.UTC()})
}

func (h *Handler) Upcoming(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		x, e := strconv.Atoi(v)
		if e != nil {
			response.Error(w, 400, "VALIDATION_ERROR", "days must be an integer")
			return
		}
		days = x
	}
	var appID *uuid.UUID
	if v := r.URL.Query().Get("applicationId"); v != "" {
		x, e := uuid.Parse(v)
		if e != nil {
			response.Error(w, 400, "INVALID_APPLICATION_ID", "applicationId must be a UUID")
			return
		}
		appID = &x
	}
	var status *Status
	if raw := r.URL.Query().Get("status"); raw != "" {
		x := Status(raw)
		if !x.Valid() {
			response.Error(w, 400, "INVALID_APPLICATION_STATUS", "status is invalid")
			return
		}
		status = &x
	}
	v, e := h.service.UpcomingDeadlines(r.Context(), days, appID, status)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]DeadlineResponse, len(v))
	for i := range v {
		out[i] = toDeadline(&v[i], h.service.now())
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	v, e := h.service.Dashboard(r.Context())
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, v)
}

func (h *Handler) RunResearch(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	q := ResearchRunRequest{}
	if r.ContentLength != 0 && !h.body(w, r, &q, true) {
		return
	}
	v, e := h.research.Run(r.Context(), a, q)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 201, toResearchRun(v))
}
func (h *Handler) ListResearch(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	v, e := h.research.ListRuns(r.Context(), a)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ResearchRunResponse, len(v))
	for i := range v {
		out[i] = toResearchRun(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) GetResearch(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "runID", "INVALID_RESEARCH_RUN_ID")
	if !ok {
		return
	}
	v, e := h.research.GetRun(r.Context(), a, id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toResearchRun(v))
}
func (h *Handler) ListSources(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "runID", "INVALID_RESEARCH_RUN_ID")
	if !ok {
		return
	}
	v, e := h.research.ListSources(r.Context(), a, id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ResearchSourceResponse, len(v))
	for i := range v {
		out[i] = toResearchSource(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) ListFindings(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	id, ok := h.id(w, r, "runID", "INVALID_RESEARCH_RUN_ID")
	if !ok {
		return
	}
	v, e := h.research.ListFindings(r.Context(), a, id)
	if e != nil {
		h.err(w, r, e)
		return
	}
	out := make([]ResearchFindingResponse, len(v))
	for i := range v {
		out[i] = toResearchFinding(&v[i])
	}
	response.Collection(w, out, len(out))
}
func (h *Handler) ReviewFinding(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	run, ok := h.id(w, r, "runID", "INVALID_RESEARCH_RUN_ID")
	if !ok {
		return
	}
	id, ok := h.id(w, r, "findingID", "INVALID_RESEARCH_FINDING_ID")
	if !ok {
		return
	}
	var q ReviewFindingRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, e := h.research.ReviewFinding(r.Context(), a, run, id, q.ReviewStatus)
	if e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, toResearchFinding(v))
}
func (h *Handler) ApplyResearch(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	run, ok := h.id(w, r, "runID", "INVALID_RESEARCH_RUN_ID")
	if !ok {
		return
	}
	if e := h.research.Apply(r.Context(), a, run); e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, map[string]string{"status": "applied"})
}
func (h *Handler) MarkResearchStale(w http.ResponseWriter, r *http.Request) {
	a, ok := h.appID(w, r)
	if !ok {
		return
	}
	if e := h.research.MarkStale(r.Context(), a); e != nil {
		h.err(w, r, e)
		return
	}
	response.Data(w, 200, map[string]string{"status": "stale"})
}

func (h *Handler) err(w http.ResponseWriter, r *http.Request, e error) {
	status, code, message := 500, "INTERNAL_ERROR", "internal server error"
	switch {
	case errors.Is(e, ErrApplicationNotFound):
		status, code, message = 404, "APPLICATION_NOT_FOUND", e.Error()
	case errors.Is(e, ErrInvalidApplicationStatus):
		status, code, message = 409, "INVALID_APPLICATION_STATUS", e.Error()
	case errors.Is(e, ErrApplicationProfileRequired), errors.Is(e, profile.ErrInvalidProfileParent):
		status, code, message = 400, "APPLICATION_PROFILE_REQUIRED", e.Error()
	case errors.Is(e, ErrRequirementNotFound):
		status, code, message = 404, "REQUIREMENT_NOT_FOUND", e.Error()
	case errors.Is(e, ErrDeadlineNotFound):
		status, code, message = 404, "DEADLINE_NOT_FOUND", e.Error()
	case errors.Is(e, ErrFundingNotFound):
		status, code, message = 404, "FUNDING_NOT_FOUND", e.Error()
	case errors.Is(e, ErrContactNotFound):
		status, code, message = 404, "CONTACT_NOT_FOUND", e.Error()
	case errors.Is(e, ErrSupervisorNotFound):
		status, code, message = 404, "SUPERVISOR_NOT_FOUND", e.Error()
	case errors.Is(e, ErrURLNotFound):
		status, code, message = 404, "URL_NOT_FOUND", e.Error()
	case errors.Is(e, ErrTaskNotFound):
		status, code, message = 404, "TASK_NOT_FOUND", e.Error()
	case errors.Is(e, ErrTaskParentCycle):
		status, code, message = 409, "TASK_PARENT_CYCLE", e.Error()
	case errors.Is(e, ErrResearchRunNotFound):
		status, code, message = 404, "RESEARCH_RUN_NOT_FOUND", e.Error()
	case errors.Is(e, ErrResearchFindingNotFound):
		status, code, message = 404, "RESEARCH_FINDING_NOT_FOUND", e.Error()
	case errors.Is(e, ErrResearchFindingConflict):
		status, code, message = 409, "RESEARCH_FINDING_CONFLICT", e.Error()
	case errors.Is(e, ErrResearchNotReady):
		status, code, message = 409, "RESEARCH_NOT_READY", e.Error()
	case errors.Is(e, ErrInvalidResearchState):
		status, code, message = 409, "INVALID_RESEARCH_STATE", e.Error()
	case errors.Is(e, catalog.ErrInstitutionNotFound), errors.Is(e, catalog.ErrProgrammeNotFound), errors.Is(e, catalog.ErrScholarshipNotFound):
		status, code, message = 400, "INVALID_CATALOG_REFERENCE", e.Error()
	case isValidationError(e):
		status, code, message = 400, "VALIDATION_ERROR", e.Error()
	}
	if status == 500 {
		appLogger.Error(r.Context(), "application request failed", "error", e)
	}
	response.Error(w, status, code, message)
}
