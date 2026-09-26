package workflow

import (
	"encoding/json"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
)

func researchTaskResponse(item *ResearchTask) ResearchTaskResponse {
	return ResearchTaskResponse{ID: item.ID, UserID: item.UserID, ParentTaskID: item.ParentTaskID, TargetApplicationID: item.TargetApplicationID, ProfileID: item.ProfileID, Title: item.Title, Description: item.Description, Instructions: item.Instructions, TaskType: item.TaskType, Status: item.Status, Priority: item.Priority, DueAt: utcTime(item.DueAt), ResearchConfig: json.RawMessage(item.ResearchConfig), StartedAt: utcTime(item.StartedAt), CompletedAt: utcTime(item.CompletedAt), FailedAt: utcTime(item.FailedAt), FailureReason: item.FailureReason, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC()}
}

func researchTaskResponses(items []ResearchTask) []ResearchTaskResponse {
	result := make([]ResearchTaskResponse, len(items))
	for i := range items {
		result[i] = researchTaskResponse(&items[i])
	}
	return result
}
func taskLinkResponse(item *ResearchTaskLink) TaskLinkResponse {
	return TaskLinkResponse{item.ID, item.ResearchTaskID, item.Label, item.URL, item.LinkType, item.CreatedAt.UTC(), item.UpdatedAt.UTC()}
}
func taskLinkResponses(items []ResearchTaskLink) []TaskLinkResponse {
	result := make([]TaskLinkResponse, len(items))
	for i := range items {
		result[i] = taskLinkResponse(&items[i])
	}
	return result
}
func taskOutputResponses(items []ResearchTaskOutput) []TaskOutputResponse {
	result := make([]TaskOutputResponse, len(items))
	for i := range items {
		result[i] = TaskOutputResponse{items[i].ID, items[i].ResearchTaskID, items[i].OutputType, items[i].EntityID, items[i].CreatedAt.UTC()}
	}
	return result
}

func proposalResponse(item *ApplicationProposal, sources []ProposalSourceRecord) ApplicationProposalResponse {
	response := ApplicationProposalResponse{ID: item.ID, UserID: item.UserID, ResearchTaskID: item.ResearchTaskID, ResearchRunID: item.ResearchRunID, InstitutionID: item.InstitutionID, ProgrammeID: item.ProgrammeID, ScholarshipID: item.ScholarshipID, ProposedInstitution: json.RawMessage(item.ProposedInstitution), ProposedProgramme: json.RawMessage(item.ProposedProgramme), ProposedScholarship: json.RawMessage(item.ProposedScholarship), Name: item.Name, Country: item.Country, Intake: item.Intake, IntakeYear: item.IntakeYear, Summary: item.Summary, Status: item.Status, Confidence: item.Confidence, ReasoningSummary: item.ReasoningSummary, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC(), ReviewedAt: utcTime(item.ReviewedAt), Sources: make([]ProposalSourceResponse, len(sources))}
	for i, source := range sources {
		response.Sources[i] = ProposalSourceResponse{ID: source.Link.ID, ResearchSourceID: source.Source.ID, SourceRole: source.Link.SourceRole, URL: source.Source.URL, Title: source.Source.Title, IsOfficial: source.Source.IsOfficial}
	}
	return response
}

func fieldResponse(item *ApplicationField) ApplicationFieldResponse {
	return ApplicationFieldResponse{item.ID, item.ApplicationID, item.Key, item.Label, json.RawMessage(item.Value), item.ValueType, item.Status, item.SourceType, item.SourceEntityID, item.PrefillRunID, item.Notes, item.CreatedAt.UTC(), item.UpdatedAt.UTC()}
}
func fieldResponses(items []ApplicationField) []ApplicationFieldResponse {
	result := make([]ApplicationFieldResponse, len(items))
	for i := range items {
		result[i] = fieldResponse(&items[i])
	}
	return result
}
func questionnaireResponse(item *ApplicationQuestionnaire, questions []ApplicationQuestion) QuestionnaireResponse {
	return QuestionnaireResponse{ID: item.ID, ApplicationID: item.ApplicationID, Title: item.Title, Description: item.Description, QuestionnaireType: item.QuestionnaireType, Status: item.Status, SourceURL: item.SourceURL, Questions: questionResponses(questions), CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC()}
}
func questionnaireResponses(items []ApplicationQuestionnaire) []QuestionnaireResponse {
	result := make([]QuestionnaireResponse, len(items))
	for i := range items {
		result[i] = questionnaireResponse(&items[i], nil)
	}
	return result
}
func questionResponse(item *ApplicationQuestion) QuestionResponse {
	return QuestionResponse{item.ID, item.QuestionnaireID, item.Key, item.Prompt, item.HelpText, item.QuestionType, item.IsRequired, item.WordLimit, item.CharacterLimit, item.SortOrder, json.RawMessage(item.Options), item.Status, item.CreatedAt.UTC(), item.UpdatedAt.UTC()}
}
func questionResponses(items []ApplicationQuestion) []QuestionResponse {
	result := make([]QuestionResponse, len(items))
	for i := range items {
		result[i] = questionResponse(&items[i])
	}
	return result
}
func answerResponse(item *ApplicationAnswer) AnswerResponse {
	return AnswerResponse{item.ID, item.QuestionID, json.RawMessage(item.Value), item.DraftText, item.Status, item.AnswerSource, item.SourceEntityType, item.SourceEntityID, item.Confidence, item.CreatedBy, item.PrefillRunID, item.AgentProvider, item.AgentModel, item.CreatedAt.UTC(), item.UpdatedAt.UTC()}
}
func answerResponses(items []ApplicationAnswer) []AnswerResponse {
	result := make([]AnswerResponse, len(items))
	for i := range items {
		result[i] = answerResponse(&items[i])
	}
	return result
}

func informationRequestResponse(item *InformationRequest, history []InformationRequestResponse) InformationRequestResponseDTO {
	response := InformationRequestResponseDTO{ID: item.ID, UserID: item.UserID, ApplicationID: item.ApplicationID, ResearchTaskID: item.ResearchTaskID, QuestionnaireID: item.QuestionnaireID, QuestionID: item.QuestionID, ApplicationFieldID: item.ApplicationFieldID, RequestType: item.RequestType, Title: item.Title, Prompt: item.Prompt, Context: item.Context, Status: item.Status, Priority: item.Priority, ResponseType: item.ResponseType, Options: json.RawMessage(item.Options), ResponseValue: json.RawMessage(item.ResponseValue), ResponseText: item.ResponseText, CreatedBy: item.CreatedBy, ResolutionSource: item.ResolutionSource, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC(), CompletedAt: utcTime(item.CompletedAt), ReopenedAt: utcTime(item.ReopenedAt), Responses: make([]InformationResponseDTO, len(history))}
	for i, previous := range history {
		response.Responses[i] = InformationResponseDTO{previous.ID, json.RawMessage(previous.ResponseValue), previous.ResponseText, previous.SubmittedBy, previous.CreatedAt.UTC()}
	}
	return response
}

func informationRequestResponses(items []InformationRequest) []InformationRequestResponseDTO {
	result := make([]InformationRequestResponseDTO, len(items))
	for i := range items {
		result[i] = informationRequestResponse(&items[i], nil)
	}
	return result
}
func prefillRunResponse(item *ApplicationPrefillRun, actions []PrefillAction) PrefillRunResponse {
	response := PrefillRunResponse{ID: item.ID, ApplicationID: item.ApplicationID, ResearchTaskID: item.ResearchTaskID, Status: item.Status, Trigger: item.Trigger, AgentProvider: item.AgentProvider, AgentModel: item.AgentModel, StartedAt: utcTime(item.StartedAt), CompletedAt: utcTime(item.CompletedAt), ErrorMessage: item.ErrorMessage, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC(), Actions: make([]PrefillActionResponse, len(actions))}
	for i, action := range actions {
		response.Actions[i] = PrefillActionResponse{action.ID, action.ActionType, action.TargetType, action.TargetID, action.Status, action.Summary, action.SourceType, action.SourceID, action.CreatedAt.UTC()}
	}
	return response
}

func applicationResponse(item *application.Application) application.ApplicationResponse {
	response := application.ApplicationResponse{ID: item.ID, UserID: item.UserID, ApplicantProfileID: item.ApplicantProfileID, Name: item.Name, Intake: item.Intake, IntakeYear: item.IntakeYear, Country: item.Country, Status: item.Status, Priority: item.Priority, Notes: item.Notes, ResearchStatus: item.ResearchStatus, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC()}
	if item.Institution != nil {
		response.Institution = &application.NamedResource{ID: item.Institution.ID, Name: item.Institution.Name}
	}
	if item.Programme != nil {
		response.Programme = &application.NamedResource{ID: item.Programme.ID, Name: item.Programme.Name}
	}
	if item.Scholarship != nil {
		response.Scholarship = &application.NamedResource{ID: item.Scholarship.ID, Name: item.Scholarship.Name}
	}
	return response
}

func researchRunDTOs(items []application.ResearchRun) []ResearchRunDTO {
	result := make([]ResearchRunDTO, len(items))
	for i, item := range items {
		result[i] = ResearchRunDTO{item.ID, item.ApplicationID, item.ResearchTaskID, item.Status, item.Trigger, item.ResearchType, item.ModelProvider, item.ModelName, utcTime(item.StartedAt), utcTime(item.CompletedAt), item.ErrorMessage, item.CreatedAt.UTC(), item.UpdatedAt.UTC()}
	}
	return result
}
func researchSourceDTOs(items []application.ResearchSource) []ResearchSourceDTO {
	result := make([]ResearchSourceDTO, len(items))
	for i, item := range items {
		result[i] = ResearchSourceDTO{item.ID, item.ResearchRunID, item.ApplicationID, item.URL, item.Title, item.Publisher, item.SourceType, item.IsOfficial, item.RetrievedAt.UTC(), item.CreatedAt.UTC()}
	}
	return result
}
func researchSourceDTO(item *application.ResearchSource) ResearchSourceDTO {
	return ResearchSourceDTO{item.ID, item.ResearchRunID, item.ApplicationID, item.URL, item.Title, item.Publisher, item.SourceType, item.IsOfficial, item.RetrievedAt.UTC(), item.CreatedAt.UTC()}
}
func researchFindingDTOs(items []application.ResearchFinding) []ResearchFindingDTO {
	result := make([]ResearchFindingDTO, len(items))
	for i, item := range items {
		result[i] = ResearchFindingDTO{item.ID, item.ResearchRunID, item.ApplicationID, item.SourceID, item.Category, item.Field, json.RawMessage(item.Value), item.Confidence, item.VerificationStatus, item.ReviewStatus, item.RawText, item.CreatedAt.UTC()}
	}
	return result
}
func researchFindingDTO(item *application.ResearchFinding) ResearchFindingDTO {
	return ResearchFindingDTO{item.ID, item.ResearchRunID, item.ApplicationID, item.SourceID, item.Category, item.Field, json.RawMessage(item.Value), item.Confidence, item.VerificationStatus, item.ReviewStatus, item.RawText, item.CreatedAt.UTC()}
}

func utcTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}
