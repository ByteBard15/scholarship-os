package workflow

import (
	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/ownership"
	appmw "github.com/byte/scholarship-os/apps/api/internal/middleware"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/go-chi/chi/v5"
)

func MountRoutes(r chi.Router, h *Handler, owners *ownership.Middleware) {
	r.Get("/research-tasks", h.ListResearchTasks)
	r.Post("/research-tasks", h.CreateResearchTask)
	r.Get("/research-tasks/export", h.ExportResearchTasks)
	r.Post("/research-tasks/import", h.ImportResearchTasks)
	r.Get("/research-contexts", h.ListResearchContexts)
	r.Post("/research-contexts", h.CreateResearchContext)
	r.Patch("/research-contexts/{contextID}", h.UpdateResearchContext)
	r.Delete("/research-contexts/{contextID}", h.DeleteResearchContext)
	r.With(owners.ResearchTask("get research task")).Get("/research-tasks/{taskID}", h.GetResearchTask)
	r.With(owners.ResearchTask("update research task")).Patch("/research-tasks/{taskID}", h.UpdateResearchTask)
	r.With(owners.ResearchTask("delete research task")).Delete("/research-tasks/{taskID}", h.DeleteResearchTask)
	r.With(owners.ResearchTask("queue research task")).Post("/research-tasks/{taskID}/queue", h.transitionTask("queue"))
	r.With(owners.ResearchTask("cancel research task")).Post("/research-tasks/{taskID}/cancel", h.transitionTask("cancel"))
	r.With(owners.ResearchTask("retry research task")).Post("/research-tasks/{taskID}/retry", h.transitionTask("retry"))
	r.With(owners.ResearchTask("restart research task")).Post("/research-tasks/{taskID}/restart", h.RestartResearchTask)
	r.With(owners.ResearchTask("list research task outputs")).Get("/research-tasks/{taskID}/outputs", h.ListTaskOutputs)
	r.With(owners.ResearchTask("list research task links")).Get("/research-tasks/{taskID}/links", h.ListTaskLinks)
	r.With(owners.ResearchTask("list research task contexts")).Get("/research-tasks/{taskID}/contexts", h.ListTaskContexts)
	r.With(owners.ResearchTask("attach research task contexts")).Post("/research-tasks/{taskID}/contexts", h.AttachTaskContexts)
	r.With(owners.ResearchTask("detach research task context")).Delete("/research-tasks/{taskID}/contexts/{contextID}", h.DetachTaskContext)
	r.With(owners.ResearchTask("create research task link")).Post("/research-tasks/{taskID}/links", h.CreateTaskLink)
	r.With(owners.ResearchTask("update research task link")).Patch("/research-tasks/{taskID}/links/{linkID}", h.UpdateTaskLink)
	r.With(owners.ResearchTask("delete research task link")).Delete("/research-tasks/{taskID}/links/{linkID}", h.DeleteTaskLink)
	r.With(owners.ResearchTask("list research task runs")).Get("/research-tasks/{taskID}/runs", h.ListTaskRuns)
	r.With(owners.ResearchTask("list research task run sources")).Get("/research-tasks/{taskID}/runs/{runID}/sources", h.ListTaskRunSources)
	r.With(owners.ResearchTask("list research task run findings")).Get("/research-tasks/{taskID}/runs/{runID}/findings", h.ListTaskRunFindings)

	r.Get("/application-proposals", h.ListProposals)
	r.With(owners.ApplicationProposal("get application proposal")).Get("/application-proposals/{proposalID}", h.GetProposal)
	r.With(owners.ApplicationProposal("approve application proposal")).Post("/application-proposals/{proposalID}/approve", h.ApproveProposal)
	r.With(owners.ApplicationProposal("reject application proposal")).Post("/application-proposals/{proposalID}/reject", h.reviewProposal("reject"))
	r.With(owners.ApplicationProposal("reopen application proposal")).Post("/application-proposals/{proposalID}/reopen", h.reviewProposal("reopen"))

	r.With(owners.Application("list application fields")).Get("/applications/{applicationID}/fields", h.ListFields)
	r.With(owners.Application("create application field")).Post("/applications/{applicationID}/fields", h.CreateField)
	r.With(owners.Application("update application field")).Patch("/applications/{applicationID}/fields/{fieldID}", h.UpdateField)
	r.With(owners.Application("delete application field")).Delete("/applications/{applicationID}/fields/{fieldID}", h.DeleteField)
	r.With(owners.Application("list application questionnaires")).Get("/applications/{applicationID}/questionnaires", h.ListQuestionnaires)
	r.With(owners.Application("create application questionnaire")).Post("/applications/{applicationID}/questionnaires", h.CreateQuestionnaire)
	r.With(owners.Application("get application questionnaire")).Get("/applications/{applicationID}/questionnaires/{questionnaireID}", h.GetQuestionnaire)
	r.With(owners.Application("update application questionnaire")).Patch("/applications/{applicationID}/questionnaires/{questionnaireID}", h.UpdateQuestionnaire)
	r.With(owners.Application("delete application questionnaire")).Delete("/applications/{applicationID}/questionnaires/{questionnaireID}", h.DeleteQuestionnaire)
	r.With(owners.Questionnaire("list questionnaire questions")).Get("/questionnaires/{questionnaireID}/questions", h.ListQuestions)
	r.With(owners.Questionnaire("create questionnaire question")).Post("/questionnaires/{questionnaireID}/questions", h.CreateQuestion)
	r.With(owners.Questionnaire("update questionnaire question")).Patch("/questionnaires/{questionnaireID}/questions/{questionID}", h.UpdateQuestion)
	r.With(owners.Questionnaire("delete questionnaire question")).Delete("/questionnaires/{questionnaireID}/questions/{questionID}", h.DeleteQuestion)
	r.With(owners.Question("list question answers")).Get("/questions/{questionID}/answers", h.ListAnswers)
	r.With(owners.Question("create question answer")).Post("/questions/{questionID}/answers", h.CreateAnswer)
	r.With(owners.Question("update question answer")).Patch("/questions/{questionID}/answers/{answerID}", h.UpdateAnswer)
	r.With(owners.Question("approve question answer")).Post("/questions/{questionID}/answers/{answerID}/approve", h.reviewAnswer("approve"))
	r.With(owners.Question("reject question answer")).Post("/questions/{questionID}/answers/{answerID}/reject", h.reviewAnswer("reject"))
	r.With(owners.Question("regenerate question answer")).Post("/questions/{questionID}/regenerate", h.RegenerateQuestion)

	r.Get("/information-requests", h.ListInformationRequests)
	r.With(owners.InformationRequest("get information request")).Get("/information-requests/{requestID}", h.GetInformationRequest)
	r.With(owners.InformationRequest("respond to information request")).Post("/information-requests/{requestID}/respond", h.RespondInformationRequest)
	r.With(owners.InformationRequest("process information request")).Post("/information-requests/{requestID}/process", h.ProcessInformationRequest)
	r.With(owners.InformationRequest("reopen information request")).Post("/information-requests/{requestID}/reopen", h.transitionInformation("reopen"))
	r.With(owners.InformationRequest("cancel information request")).Post("/information-requests/{requestID}/cancel", h.transitionInformation("cancel"))

	r.With(owners.Application("run application prefill")).Post("/applications/{applicationID}/prefill", h.RunPrefill)
	r.With(owners.Application("get application preparation summary")).Get("/applications/{applicationID}/preparation", h.PreparationSummary)
}

func MountAgentRoutes(r chi.Router, h *Handler, applications *application.Handler, owners *ownership.Middleware) {
	research := appmw.RequireAgentScope(principal.ScopeResearch)
	applicationScope := appmw.RequireAgentScope(principal.ScopeApplications)
	profileScope := appmw.RequireAgentScope(principal.ScopeProfiles)
	informationScope := appmw.RequireAgentScope(principal.ScopeInformationRequests)

	r.With(research).Get("/agent/research-tasks", h.PollResearchTasks)
	r.With(research, owners.ResearchTask("agent get research task")).Get("/agent/research-tasks/{taskID}", h.GetResearchTask)
	r.With(research, owners.ResearchTask("agent list research task links")).Get("/agent/research-tasks/{taskID}/links", h.ListTaskLinks)
	r.With(research, owners.ResearchTask("agent list research task contexts")).Get("/agent/research-tasks/{taskID}/contexts", h.ListTaskContexts)
	r.With(research, owners.ResearchTask("agent list research task outputs")).Get("/agent/research-tasks/{taskID}/outputs", h.ListTaskOutputs)
	r.With(research, owners.ResearchTask("agent get research task effective profile")).Get("/agent/research-tasks/{taskID}/effective-profile", h.AgentGetResearchTaskEffectiveProfile)
	r.With(research, owners.ResearchTask("agent start research task")).Post("/agent/research-tasks/{taskID}/start", h.StartResearchTask)
	r.With(research, owners.ResearchTask("agent list research task runs")).Get("/agent/research-tasks/{taskID}/runs", h.ListTaskRuns)
	r.With(research, owners.ResearchTask("agent create research source")).Post("/agent/research-tasks/{taskID}/sources", h.AgentCreateResearchSource)
	r.With(research, owners.ResearchTask("agent create research finding")).Post("/agent/research-tasks/{taskID}/findings", h.AgentCreateResearchFinding)
	r.With(research, owners.ResearchTask("agent complete research task")).Post("/agent/research-tasks/{taskID}/complete", h.CompleteResearchTask)
	r.With(research).Get("/agent/research-contexts", h.ListResearchContexts)
	r.With(research).Post("/agent/application-proposals", h.AgentCreateApplicationProposal)
	r.With(applicationScope).Get("/agent/application-proposals/approved", h.AgentListApprovedProposals)
	r.With(applicationScope, owners.ApplicationProposal("agent create application from approved proposal")).Post("/agent/application-proposals/{proposalID}/application", h.AgentCreateApplicationFromProposal)

	r.With(applicationScope, owners.Application("agent get application")).Get("/agent/applications/{applicationID}", applications.Get)
	r.With(profileScope, owners.Application("agent get application effective profile")).Get("/agent/applications/{applicationID}/effective-profile", h.AgentGetEffectiveProfile)
	r.With(applicationScope, owners.Application("agent run application prefill")).Post("/agent/applications/{applicationID}/prefill", h.RunPrefill)

	r.With(informationScope).Post("/agent/information-requests", h.AgentCreateInformationRequest)
	r.With(informationScope, owners.InformationRequest("agent get information request")).Get("/agent/information-requests/{requestID}", h.GetInformationRequest)
	r.With(informationScope, owners.InformationRequest("agent process information request")).Post("/agent/information-requests/{requestID}/process", h.ProcessInformationRequest)
}
