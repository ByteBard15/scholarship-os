package workflow

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
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
		response.Error(w, 400, "INVALID_JSON", "request body must be valid JSON")
		return false
	}
	if err := h.validate.Struct(value); err != nil {
		response.Error(w, 400, "VALIDATION_ERROR", err.Error())
		return false
	}
	return true
}
func (h *Handler) id(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		response.Error(w, 400, "INVALID_ID", key+" must be a UUID")
		return uuid.Nil, false
	}
	return value, true
}

func (h *Handler) CreateResearchTask(w http.ResponseWriter, r *http.Request) {
	var request CreateResearchTaskRequest
	if !h.decode(w, r, &request) {
		return
	}
	task, err := h.service.CreateResearchTask(r.Context(), request, r.URL.Query().Get("queue") == "true")
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, researchTaskResponse(task))
}

func researchContextResponse(item *ResearchContext) ResearchContextResponse {
	return ResearchContextResponse{ID: item.ID, UserID: item.UserID, Question: item.Question, Answer: item.Answer, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC()}
}

func researchContextResponses(items []ResearchContext) []ResearchContextResponse {
	result := make([]ResearchContextResponse, len(items))
	for i := range items {
		result[i] = researchContextResponse(&items[i])
	}
	return result
}

func (h *Handler) ListResearchContexts(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if raw := r.URL.Query().Get("userId"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "userId must be a UUID")
			return
		}
		userID = &parsed
	}
	items, err := h.service.ListResearchContexts(r.Context(), userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchContextResponses(items)
	response.Collection(w, result, len(result))
}

func (h *Handler) CreateResearchContext(w http.ResponseWriter, r *http.Request) {
	var request CreateResearchContextRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CreateResearchContext(r.Context(), request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, researchContextResponse(item))
}

func (h *Handler) UpdateResearchContext(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "contextID")
	if !ok {
		return
	}
	var request NewResearchContextRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateResearchContext(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, researchContextResponse(item))
}

func (h *Handler) DeleteResearchContext(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "contextID")
	if !ok {
		return
	}
	if err := h.service.DeleteResearchContext(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTaskContexts(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	items, err := h.service.ListTaskContexts(r.Context(), taskID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchContextResponses(items)
	response.Collection(w, result, len(result))
}

func (h *Handler) AttachTaskContexts(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var request AttachResearchContextsRequest
	if !h.decode(w, r, &request) {
		return
	}
	items, err := h.service.AttachTaskContexts(r.Context(), taskID, request.ResearchContextIDs)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchContextResponses(items)
	response.Collection(w, result, len(result))
}

func (h *Handler) DetachTaskContext(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	contextID, ok := h.id(w, r, "contextID")
	if !ok {
		return
	}
	if err := h.service.DetachTaskContext(r.Context(), taskID, contextID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AgentCreateResearchSource(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var request AgentResearchSourceRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.CreateAgentResearchSource(r.Context(), taskID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, researchSourceDTO(value))
}

func (h *Handler) AgentCreateResearchFinding(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var request AgentResearchFindingRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.CreateAgentResearchFinding(r.Context(), taskID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, researchFindingDTO(value))
}

func (h *Handler) AgentCreateApplicationProposal(w http.ResponseWriter, r *http.Request) {
	var request AgentApplicationProposalRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.CreateAgentApplicationProposal(r.Context(), request.ResearchTaskID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, value)
}

func (h *Handler) AgentGetEffectiveProfile(w http.ResponseWriter, r *http.Request) {
	applicationID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	value, err := h.service.GetApplicationEffectiveProfile(r.Context(), applicationID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, value)
}

func (h *Handler) AgentCreateInformationRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal.PrincipalFromContext(r.Context())
	if !ok || actor.UserID == nil {
		response.Error(w, http.StatusForbidden, "FORBIDDEN", "agent credential must belong to a user")
		return
	}
	var request AgentInformationRequestRequest
	if !h.decode(w, r, &request) {
		return
	}
	value, err := h.service.CreateInformationRequest(r.Context(), CreateInformationRequest{UserID: *actor.UserID, ApplicationID: request.ApplicationID, ResearchTaskID: request.ResearchTaskID, QuestionnaireID: request.QuestionnaireID, QuestionID: request.QuestionID, ApplicationFieldID: request.ApplicationFieldID, RequestType: request.RequestType, Title: request.Title, Prompt: request.Prompt, Context: request.Context, Priority: request.Priority, ResponseType: request.ResponseType, Options: request.Options, CreatedBy: "agent"})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, informationRequestResponse(value, nil))
}
func (h *Handler) ListResearchTasks(w http.ResponseWriter, r *http.Request) {
	filters := ResearchTaskFilters{}
	if raw := r.URL.Query().Get("userId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, 400, "INVALID_USER_ID", "userId must be a UUID")
			return
		}
		filters.UserID = &id
	}
	if value := r.URL.Query().Get("status"); value != "" {
		filters.Status = &value
	}
	if value := r.URL.Query().Get("taskType"); value != "" {
		filters.TaskType = &value
	}
	if value := r.URL.Query().Get("priority"); value != "" {
		filters.Priority = &value
	}
	items, err := h.service.ListResearchTasks(r.Context(), filters)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchTaskResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) PollResearchTasks(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 50 {
			response.Error(w, http.StatusBadRequest, "INVALID_LIMIT", "limit must be between 1 and 50")
			return
		}
		limit = parsed
	}
	status := "queued"
	items, err := h.service.ListResearchTasks(r.Context(), ResearchTaskFilters{Status: &status, Limit: limit, OldestFirst: true})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchTaskResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) GetResearchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	item, err := h.service.GetResearchTask(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, researchTaskResponse(item))
}
func (h *Handler) UpdateResearchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var request UpdateResearchTaskRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateResearchTask(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, researchTaskResponse(item))
}
func (h *Handler) DeleteResearchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	if err := h.service.DeleteResearchTask(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) transitionTask(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.id(w, r, "taskID")
		if !ok {
			return
		}
		item, err := h.service.TransitionResearchTask(r.Context(), id, action)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 200, researchTaskResponse(item))
	}
}
func (h *Handler) StartResearchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	item, run, err := h.service.StartResearchTask(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	runDTO := researchRunDTOs([]application.ResearchRun{*run})[0]
	response.Data(w, 200, ClaimedResearchTaskResponse{Task: researchTaskResponse(item), ResearchRun: runDTO})
}
func (h *Handler) CompleteResearchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var request CompleteAgentResearchRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CompleteResearchTask(r.Context(), id, request.ResearchRunID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, http.StatusOK, researchTaskResponse(item))
}
func (h *Handler) ListTaskOutputs(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	items, err := h.service.ListTaskOutputs(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := taskOutputResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) ListTaskRuns(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	items, err := h.service.ListTaskRuns(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchRunDTOs(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) ListTaskRunSources(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	runID, ok := h.id(w, r, "runID")
	if !ok {
		return
	}
	items, err := h.service.ListRunSources(r.Context(), taskID, runID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchSourceDTOs(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) ListTaskRunFindings(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	runID, ok := h.id(w, r, "runID")
	if !ok {
		return
	}
	items, err := h.service.ListRunFindings(r.Context(), taskID, runID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := researchFindingDTOs(items)
	response.Collection(w, result, len(result))
}

func (h *Handler) ListTaskLinks(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	items, err := h.service.ListTaskLinks(r.Context(), taskID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := taskLinkResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) CreateTaskLink(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	var request TaskLinkRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CreateTaskLink(r.Context(), taskID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, taskLinkResponse(item))
}
func (h *Handler) UpdateTaskLink(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	linkID, ok := h.id(w, r, "linkID")
	if !ok {
		return
	}
	var request TaskLinkRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateTaskLink(r.Context(), taskID, linkID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, taskLinkResponse(item))
}
func (h *Handler) DeleteTaskLink(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.id(w, r, "taskID")
	if !ok {
		return
	}
	linkID, ok := h.id(w, r, "linkID")
	if !ok {
		return
	}
	if err := h.service.DeleteTaskLink(r.Context(), taskID, linkID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListProposals(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if raw := r.URL.Query().Get("userId"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, 400, "INVALID_USER_ID", "userId must be a UUID")
			return
		}
		userID = &parsed
	}
	var status *string
	if value := r.URL.Query().Get("status"); value != "" {
		status = &value
	}
	result, err := h.service.ListProposalResponses(r.Context(), userID, status)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Collection(w, result, len(result))
}
func (h *Handler) GetProposal(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "proposalID")
	if !ok {
		return
	}
	item, err := h.service.GetProposal(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, item)
}
func (h *Handler) ApproveProposal(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "proposalID")
	if !ok {
		return
	}
	var request ApproveProposalRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.ApproveProposal(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, applicationResponse(item))
}
func (h *Handler) reviewProposal(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.id(w, r, "proposalID")
		if !ok {
			return
		}
		_, err := h.service.ReviewProposal(r.Context(), id, action)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		item, err := h.service.GetProposal(r.Context(), id)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 200, item)
	}
}

func (h *Handler) ListFields(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	items, err := h.service.ListFields(r.Context(), appID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := fieldResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) CreateField(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	var request ApplicationFieldRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CreateField(r.Context(), appID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, fieldResponse(item))
}
func (h *Handler) UpdateField(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	fieldID, ok := h.id(w, r, "fieldID")
	if !ok {
		return
	}
	var request ApplicationFieldRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateField(r.Context(), appID, fieldID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, fieldResponse(item))
}
func (h *Handler) DeleteField(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	fieldID, ok := h.id(w, r, "fieldID")
	if !ok {
		return
	}
	if err := h.service.DeleteField(r.Context(), appID, fieldID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListQuestionnaires(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	items, err := h.service.ListQuestionnaires(r.Context(), appID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := questionnaireResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) GetQuestionnaire(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	id, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	item, err := h.service.GetQuestionnaire(r.Context(), appID, id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, item)
}
func (h *Handler) CreateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	var request QuestionnaireRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CreateQuestionnaire(r.Context(), appID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, questionnaireResponse(item, nil))
}
func (h *Handler) UpdateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	id, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	var request QuestionnaireRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateQuestionnaire(r.Context(), appID, id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, questionnaireResponse(item, nil))
}
func (h *Handler) DeleteQuestionnaire(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	id, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	if err := h.service.DeleteQuestionnaire(r.Context(), appID, id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	items, err := h.service.ListQuestions(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := questionResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	var request QuestionRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CreateQuestion(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, questionResponse(item))
}
func (h *Handler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionnaireID, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	questionID, ok := h.id(w, r, "questionID")
	if !ok {
		return
	}
	var request QuestionRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateQuestion(r.Context(), questionnaireID, questionID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, questionResponse(item))
}
func (h *Handler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionnaireID, ok := h.id(w, r, "questionnaireID")
	if !ok {
		return
	}
	questionID, ok := h.id(w, r, "questionID")
	if !ok {
		return
	}
	if err := h.service.DeleteQuestion(r.Context(), questionnaireID, questionID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) ListAnswers(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "questionID")
	if !ok {
		return
	}
	items, err := h.service.ListAnswers(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := answerResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) CreateAnswer(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "questionID")
	if !ok {
		return
	}
	var request AnswerRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.CreateAnswer(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, answerResponse(item))
}
func (h *Handler) UpdateAnswer(w http.ResponseWriter, r *http.Request) {
	questionID, ok := h.id(w, r, "questionID")
	if !ok {
		return
	}
	answerID, ok := h.id(w, r, "answerID")
	if !ok {
		return
	}
	var request AnswerRequest
	if !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.UpdateAnswer(r.Context(), questionID, answerID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, answerResponse(item))
}
func (h *Handler) reviewAnswer(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID, ok := h.id(w, r, "questionID")
		if !ok {
			return
		}
		answerID, ok := h.id(w, r, "answerID")
		if !ok {
			return
		}
		item, err := h.service.ReviewAnswer(r.Context(), questionID, answerID, action)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 200, answerResponse(item))
	}
}
func (h *Handler) RegenerateQuestion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "questionID")
	if !ok {
		return
	}
	item, err := h.service.RegenerateQuestion(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, prefillRunResponse(item, nil))
}

func (h *Handler) ListInformationRequests(w http.ResponseWriter, r *http.Request) {
	filters := InformationRequestFilters{}
	if raw := r.URL.Query().Get("userId"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, 400, "INVALID_USER_ID", "userId must be a UUID")
			return
		}
		filters.UserID = &parsed
	}
	if raw := r.URL.Query().Get("applicationId"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, 400, "INVALID_APPLICATION_ID", "applicationId must be a UUID")
			return
		}
		filters.ApplicationID = &parsed
	}
	if value := r.URL.Query().Get("status"); value != "" {
		filters.Status = &value
	}
	if value := r.URL.Query().Get("requestType"); value != "" {
		filters.RequestType = &value
	}
	items, err := h.service.ListInformationRequests(r.Context(), filters)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	result := informationRequestResponses(items)
	response.Collection(w, result, len(result))
}
func (h *Handler) GetInformationRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "requestID")
	if !ok {
		return
	}
	item, err := h.service.GetInformationRequest(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, item)
}
func (h *Handler) RespondInformationRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "requestID")
	if !ok {
		return
	}
	var request RespondInformationRequest
	if !h.decode(w, r, &request) {
		return
	}
	_, err := h.service.RespondInformationRequest(r.Context(), id, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	item, err := h.service.GetInformationRequest(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, item)
}
func (h *Handler) ProcessInformationRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "requestID")
	if !ok {
		return
	}
	_, err := h.service.ProcessInformationRequest(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	item, err := h.service.GetInformationRequest(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, item)
}
func (h *Handler) transitionInformation(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.id(w, r, "requestID")
		if !ok {
			return
		}
		item, err := h.service.TransitionInformationRequest(r.Context(), id, action)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		response.Data(w, 200, informationRequestResponse(item, nil))
	}
}

func (h *Handler) RunPrefill(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	request := PrefillRequest{}
	if r.ContentLength != 0 && !h.decode(w, r, &request) {
		return
	}
	item, err := h.service.RunPrefill(r.Context(), appID, request)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 201, prefillRunResponse(item, nil))
}
func (h *Handler) PreparationSummary(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.id(w, r, "applicationID")
	if !ok {
		return
	}
	item, err := h.service.PreparationSummary(r.Context(), appID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	response.Data(w, 200, item)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "internal server error"
	switch {
	case errors.Is(err, ErrResearchTaskNotFound):
		status, code, message = 404, "RESEARCH_TASK_NOT_FOUND", err.Error()
	case errors.Is(err, ErrResearchTaskLinkNotFound):
		status, code, message = 404, "RESEARCH_TASK_LINK_NOT_FOUND", err.Error()
	case errors.Is(err, ErrResearchContextNotFound):
		status, code, message = 404, "RESEARCH_CONTEXT_NOT_FOUND", err.Error()
	case errors.Is(err, ErrInvalidResearchTaskState):
		status, code, message = 409, "INVALID_RESEARCH_TASK_STATE", err.Error()
	case errors.Is(err, ErrResearchTaskParentCycle):
		status, code, message = 409, "RESEARCH_TASK_PARENT_CYCLE", err.Error()
	case errors.Is(err, ErrInvalidResearchTaskParent):
		status, code, message = 400, "INVALID_RESEARCH_TASK_PARENT", err.Error()
	case errors.Is(err, ErrApplicationProposalNotFound):
		status, code, message = 404, "APPLICATION_PROPOSAL_NOT_FOUND", err.Error()
	case errors.Is(err, ErrInvalidProposalState):
		status, code, message = 409, "INVALID_APPLICATION_PROPOSAL_STATE", err.Error()
	case errors.Is(err, ErrProposalParentRequired):
		status, code, message = 400, "PROPOSAL_PARENT_PROFILE_REQUIRED", err.Error()
	case errors.Is(err, ErrApplicationFieldNotFound):
		status, code, message = 404, "APPLICATION_FIELD_NOT_FOUND", err.Error()
	case errors.Is(err, ErrQuestionnaireNotFound):
		status, code, message = 404, "QUESTIONNAIRE_NOT_FOUND", err.Error()
	case errors.Is(err, ErrQuestionNotFound):
		status, code, message = 404, "QUESTION_NOT_FOUND", err.Error()
	case errors.Is(err, ErrAnswerNotFound):
		status, code, message = 404, "ANSWER_NOT_FOUND", err.Error()
	case errors.Is(err, ErrInformationRequestNotFound):
		status, code, message = 404, "INFORMATION_REQUEST_NOT_FOUND", err.Error()
	case errors.Is(err, ErrInvalidInformationState):
		status, code, message = 409, "INVALID_INFORMATION_REQUEST_STATE", err.Error()
	case errors.Is(err, ErrInformationTargetRequired):
		status, code, message = 400, "INFORMATION_REQUEST_TARGET_REQUIRED", err.Error()
	case errors.Is(err, application.ErrApplicationNotFound):
		status, code, message = 404, "APPLICATION_NOT_FOUND", err.Error()
	case errors.As(err, new(*ValidationError)):
		status, code, message = 400, "VALIDATION_ERROR", err.Error()
	}
	if status == 500 {
		appLogger.Error(r.Context(), "workflow request failed", "error", err, "path", r.URL.Path)
	}
	response.Error(w, status, code, message)
}

func boolQuery(value string) bool { return strings.EqualFold(value, "true") }
