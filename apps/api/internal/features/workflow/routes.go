package workflow

import (
	"github.com/example/scholarship-os/apps/api/internal/features/application"
	appmw "github.com/example/scholarship-os/apps/api/internal/middleware"
	principal "github.com/example/scholarship-os/apps/api/pkg/auth"
	"github.com/go-chi/chi/v5"
)

func MountRoutes(r chi.Router, h *Handler) {
	r.Get("/research-tasks", h.ListResearchTasks)
	r.Post("/research-tasks", h.CreateResearchTask)
	r.Get("/research-tasks/{taskID}", h.GetResearchTask)
	r.Patch("/research-tasks/{taskID}", h.UpdateResearchTask)
	r.Delete("/research-tasks/{taskID}", h.DeleteResearchTask)
	r.Post("/research-tasks/{taskID}/queue", h.transitionTask("queue"))
	r.Post("/research-tasks/{taskID}/start", h.StartResearchTask)
	r.Post("/research-tasks/{taskID}/cancel", h.transitionTask("cancel"))
	r.Post("/research-tasks/{taskID}/retry", h.transitionTask("retry"))
	r.Get("/research-tasks/{taskID}/outputs", h.ListTaskOutputs)
	r.Get("/research-tasks/{taskID}/links", h.ListTaskLinks)
	r.Post("/research-tasks/{taskID}/links", h.CreateTaskLink)
	r.Patch("/research-tasks/{taskID}/links/{linkID}", h.UpdateTaskLink)
	r.Delete("/research-tasks/{taskID}/links/{linkID}", h.DeleteTaskLink)
	r.Get("/research-tasks/{taskID}/runs", h.ListTaskRuns)
	r.Get("/research-tasks/{taskID}/runs/{runID}/sources", h.ListTaskRunSources)
	r.Get("/research-tasks/{taskID}/runs/{runID}/findings", h.ListTaskRunFindings)

	r.Get("/application-proposals", h.ListProposals)
	r.Get("/application-proposals/{proposalID}", h.GetProposal)
	r.Post("/application-proposals/{proposalID}/approve", h.ApproveProposal)
	r.Post("/application-proposals/{proposalID}/reject", h.reviewProposal("reject"))
	r.Post("/application-proposals/{proposalID}/reopen", h.reviewProposal("reopen"))

	r.Get("/applications/{applicationID}/fields", h.ListFields)
	r.Post("/applications/{applicationID}/fields", h.CreateField)
	r.Patch("/applications/{applicationID}/fields/{fieldID}", h.UpdateField)
	r.Delete("/applications/{applicationID}/fields/{fieldID}", h.DeleteField)
	r.Get("/applications/{applicationID}/questionnaires", h.ListQuestionnaires)
	r.Post("/applications/{applicationID}/questionnaires", h.CreateQuestionnaire)
	r.Get("/applications/{applicationID}/questionnaires/{questionnaireID}", h.GetQuestionnaire)
	r.Patch("/applications/{applicationID}/questionnaires/{questionnaireID}", h.UpdateQuestionnaire)
	r.Delete("/applications/{applicationID}/questionnaires/{questionnaireID}", h.DeleteQuestionnaire)
	r.Get("/questionnaires/{questionnaireID}/questions", h.ListQuestions)
	r.Post("/questionnaires/{questionnaireID}/questions", h.CreateQuestion)
	r.Patch("/questionnaires/{questionnaireID}/questions/{questionID}", h.UpdateQuestion)
	r.Delete("/questionnaires/{questionnaireID}/questions/{questionID}", h.DeleteQuestion)
	r.Get("/questions/{questionID}/answers", h.ListAnswers)
	r.Post("/questions/{questionID}/answers", h.CreateAnswer)
	r.Patch("/questions/{questionID}/answers/{answerID}", h.UpdateAnswer)
	r.Post("/questions/{questionID}/answers/{answerID}/approve", h.reviewAnswer("approve"))
	r.Post("/questions/{questionID}/answers/{answerID}/reject", h.reviewAnswer("reject"))
	r.Post("/questions/{questionID}/regenerate", h.RegenerateQuestion)

	r.Get("/information-requests", h.ListInformationRequests)
	r.Get("/information-requests/{requestID}", h.GetInformationRequest)
	r.Post("/information-requests/{requestID}/respond", h.RespondInformationRequest)
	r.Post("/information-requests/{requestID}/process", h.ProcessInformationRequest)
	r.Post("/information-requests/{requestID}/reopen", h.transitionInformation("reopen"))
	r.Post("/information-requests/{requestID}/cancel", h.transitionInformation("cancel"))

	r.Post("/applications/{applicationID}/prefill", h.RunPrefill)
	r.Get("/applications/{applicationID}/preparation", h.PreparationSummary)
}

func MountAgentRoutes(r chi.Router, h *Handler, applications *application.Handler) {
	research := appmw.RequireAgentScope(principal.ScopeResearch)
	applicationScope := appmw.RequireAgentScope(principal.ScopeApplications)
	profileScope := appmw.RequireAgentScope(principal.ScopeProfiles)
	informationScope := appmw.RequireAgentScope(principal.ScopeInformationRequests)

	r.With(research).Get("/agent/research-tasks/{taskID}", h.GetResearchTask)
	r.With(research).Get("/agent/research-tasks/{taskID}/links", h.ListTaskLinks)
	r.With(research).Post("/agent/research-tasks/{taskID}/start", h.StartResearchTask)
	r.With(research).Post("/agent/research-tasks/{taskID}/sources", h.AgentCreateResearchSource)
	r.With(research).Post("/agent/research-tasks/{taskID}/findings", h.AgentCreateResearchFinding)
	r.With(research).Post("/agent/application-proposals", h.AgentCreateApplicationProposal)

	r.With(applicationScope).Get("/agent/applications/{applicationID}", applications.Get)
	r.With(profileScope).Get("/agent/applications/{applicationID}/effective-profile", h.AgentGetEffectiveProfile)
	r.With(applicationScope).Post("/agent/applications/{applicationID}/prefill", h.RunPrefill)

	r.With(informationScope).Post("/agent/information-requests", h.AgentCreateInformationRequest)
	r.With(informationScope).Get("/agent/information-requests/{requestID}", h.GetInformationRequest)
	r.With(informationScope).Post("/agent/information-requests/{requestID}/process", h.ProcessInformationRequest)
}
