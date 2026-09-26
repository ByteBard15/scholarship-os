package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	appLogger "github.com/byte/scholarship-os/apps/api/pkg/logger"
	"github.com/google/uuid"
)

type ApplicationReader interface {
	Get(context.Context, uuid.UUID) (*application.Application, error)
}

type ProfileReader interface {
	Get(context.Context, uuid.UUID) (*profile.ApplicantProfile, error)
	ResolveEffectiveProfile(context.Context, uuid.UUID) (*profile.EffectiveProfileResponse, error)
}

type Service struct {
	repo         Repository
	users        user.Repository
	applications ApplicationReader
	profiles     ProfileReader
	researcher   ResearchExecutor
	prefiller    ApplicationPrefiller
	now          func() time.Time
}

func NewService(repo Repository, users user.Repository, applications ApplicationReader, profiles ProfileReader, researcher ResearchExecutor, prefiller ApplicationPrefiller) *Service {
	return &Service{repo: repo, users: users, applications: applications, profiles: profiles, researcher: researcher, prefiller: prefiller, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) CreateResearchTask(ctx context.Context, request CreateResearchTaskRequest, queue bool) (*ResearchTask, error) {
	if principal.EnforceOwner(ctx, request.UserID) != nil {
		return nil, ErrResearchTaskNotFound
	}
	if _, err := s.users.GetByID(ctx, request.UserID); err != nil {
		return nil, err
	}
	task := &ResearchTask{UserID: request.UserID, ParentTaskID: request.ParentTaskID, TargetApplicationID: request.TargetApplicationID, Title: strings.TrimSpace(request.Title), Description: request.Description, Instructions: request.Instructions, TaskType: request.TaskType, Status: "draft", Priority: request.Priority, DueAt: request.DueAt, ResearchConfig: profile.JSON(request.ResearchConfig)}
	if queue {
		task.Status = "queued"
	}
	if err := s.validateTaskRelationships(ctx, task); err != nil {
		return nil, err
	}
	links := make([]ResearchTaskLink, len(request.Links))
	for i, link := range request.Links {
		links[i] = ResearchTaskLink{Label: link.Label, URL: link.URL, LinkType: link.LinkType}
	}
	if err := s.repo.CreateResearchTask(ctx, task, links); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) ListResearchTasks(ctx context.Context, filters ResearchTaskFilters) ([]ResearchTask, error) {
	if actor, ok := principal.PrincipalFromContext(ctx); ok && !actor.IsSystem() {
		if actor.UserID == nil {
			return nil, principal.ErrForbidden
		}
		filters.UserID = actor.UserID
	}
	return s.repo.ListResearchTasks(ctx, filters)
}
func (s *Service) GetResearchTask(ctx context.Context, id uuid.UUID) (*ResearchTask, error) {
	value, err := s.repo.GetResearchTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, value.UserID) != nil {
		return nil, ErrResearchTaskNotFound
	}
	return value, nil
}

func (s *Service) UpdateResearchTask(ctx context.Context, id uuid.UUID, request UpdateResearchTaskRequest) (*ResearchTask, error) {
	task, err := s.repo.GetResearchTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if task.Status == "running" {
		return nil, ErrInvalidResearchTaskState
	}
	if request.Title != nil {
		task.Title = strings.TrimSpace(*request.Title)
	}
	if request.Description != nil {
		task.Description = request.Description
	}
	if request.Instructions != nil {
		task.Instructions = request.Instructions
	}
	if request.TaskType != nil {
		task.TaskType = *request.TaskType
	}
	if request.ParentTaskID != nil {
		task.ParentTaskID = request.ParentTaskID
	}
	if request.TargetApplicationID != nil {
		task.TargetApplicationID = request.TargetApplicationID
	}
	if request.Priority != nil {
		task.Priority = request.Priority
	}
	if request.DueAt != nil {
		task.DueAt = request.DueAt
	}
	if request.ResearchConfig != nil {
		task.ResearchConfig = profile.JSON(request.ResearchConfig)
	}
	if err = s.validateTaskRelationships(ctx, task); err != nil {
		return nil, err
	}
	return task, s.repo.UpdateResearchTask(ctx, task)
}

func (s *Service) DeleteResearchTask(ctx context.Context, id uuid.UUID) error {
	task, err := s.repo.GetResearchTask(ctx, id)
	if err != nil {
		return err
	}
	if task.Status == "running" {
		return ErrInvalidResearchTaskState
	}
	return s.repo.DeleteResearchTask(ctx, task)
}

func (s *Service) validateTaskRelationships(ctx context.Context, task *ResearchTask) error {
	if task.ParentTaskID != nil {
		if *task.ParentTaskID == task.ID {
			return ErrResearchTaskParentCycle
		}
		parent, err := s.repo.GetResearchTask(ctx, *task.ParentTaskID)
		if err != nil || parent.UserID != task.UserID {
			return ErrInvalidResearchTaskParent
		}
		seen := map[uuid.UUID]bool{task.ID: true}
		current := parent
		for depth := 0; current != nil; depth++ {
			if depth >= 12 || seen[current.ID] {
				return ErrResearchTaskParentCycle
			}
			seen[current.ID] = true
			if current.ParentTaskID == nil {
				break
			}
			current, err = s.repo.GetResearchTask(ctx, *current.ParentTaskID)
			if err != nil {
				return ErrInvalidResearchTaskParent
			}
		}
	}
	if task.TargetApplicationID != nil {
		app, err := s.applications.Get(ctx, *task.TargetApplicationID)
		if err != nil {
			return err
		}
		if app.UserID != task.UserID {
			return validationError("target application belongs to another user")
		}
	}
	return nil
}

func (s *Service) TransitionResearchTask(ctx context.Context, id uuid.UUID, action string) (*ResearchTask, error) {
	task, err := s.repo.GetResearchTask(ctx, id)
	if err != nil {
		return nil, err
	}
	now := s.now()
	switch action {
	case "queue":
		if task.Status != "draft" && task.Status != "ready" {
			return nil, ErrInvalidResearchTaskState
		}
		task.Status = "queued"
	case "cancel":
		if task.Status == "completed" || task.Status == "cancelled" {
			return nil, ErrInvalidResearchTaskState
		}
		task.Status = "cancelled"
	case "retry":
		if task.Status != "failed" {
			return nil, ErrInvalidResearchTaskState
		}
		task.Status, task.FailedAt, task.FailureReason = "queued", nil, nil
	case "complete":
		task.Status, task.CompletedAt = "completed", &now
	default:
		return nil, ErrInvalidResearchTaskState
	}
	return task, s.repo.UpdateResearchTask(ctx, task)
}

func (s *Service) StartResearchTask(ctx context.Context, id uuid.UUID) (*ResearchTask, error) {
	task, err := s.repo.GetResearchTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if task.Status != "queued" && task.Status != "ready" && task.Status != "draft" {
		return nil, ErrInvalidResearchTaskState
	}
	now := s.now()
	task.Status, task.StartedAt, task.CompletedAt, task.FailedAt, task.FailureReason = "running", &now, nil, nil, nil
	if err = s.repo.UpdateResearchTask(ctx, task); err != nil {
		return nil, err
	}
	links, err := s.repo.ListTaskLinks(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	input := ResearchTaskInput{Task: researchTaskResponse(task), Links: taskLinkResponses(links)}
	if task.TargetApplicationID != nil {
		app, appErr := s.applications.Get(ctx, *task.TargetApplicationID)
		if appErr != nil {
			return nil, appErr
		}
		response := applicationResponse(app)
		input.TargetApplication = &response
		effective, profileErr := s.profiles.ResolveEffectiveProfile(ctx, app.ApplicantProfileID)
		if profileErr == nil {
			input.EffectiveProfile = effective
		}
	}
	provider, model := s.researcher.ProviderName(), s.researcher.ModelName()
	run := &application.ResearchRun{Base: application.Base{ID: uuid.New()}, ResearchTaskID: &task.ID, ApplicationID: task.TargetApplicationID, Status: "running", Trigger: "agent", ResearchType: task.TaskType, ModelProvider: &provider, ModelName: &model, StartedAt: &now}
	result, executeErr := s.researcher.Execute(ctx, input)
	if executeErr != nil {
		task.Status, task.FailedAt = "failed", &now
		message := executeErr.Error()
		task.FailureReason = &message
		_ = s.repo.FailResearchTask(ctx, task, run, executeErr)
		return nil, executeErr
	}
	persistence, err := s.buildResearchPersistence(task, run, result)
	if err != nil {
		return nil, err
	}
	run.Status = "review_required"
	task.Status = "review_required"
	if len(persistence.Proposals) == 0 && len(persistence.Findings) == 0 {
		task.Status, task.CompletedAt = "completed", &now
		run.Status, run.CompletedAt = "complete", &now
	}
	if err = s.repo.PersistResearchTaskResult(ctx, task, persistence); err != nil {
		return nil, err
	}
	appLogger.Info(ctx, "research task completed", "research_task_id", task.ID, "run_id", run.ID, "proposals", len(persistence.Proposals), "findings", len(persistence.Findings))
	return task, nil
}

func (s *Service) buildResearchPersistence(task *ResearchTask, run *application.ResearchRun, result *ResearchTaskResult) (ResearchPersistence, error) {
	persistence := ResearchPersistence{Run: run}
	sourceIDs := make(map[string]uuid.UUID, len(result.Sources))
	for _, candidate := range result.Sources {
		id := uuid.New()
		sourceIDs[candidate.Key] = id
		persistence.Sources = append(persistence.Sources, application.ResearchSource{ID: id, ResearchRunID: run.ID, ApplicationID: task.TargetApplicationID, URL: candidate.URL, Title: candidate.Title, Publisher: candidate.Publisher, SourceType: candidate.SourceType, IsOfficial: candidate.IsOfficial, RetrievedAt: candidate.RetrievedAt, PublishedAt: candidate.PublishedAt, ContentHash: candidate.ContentHash, Notes: candidate.Notes})
	}
	addFinding := func(sourceKey, category, field string, value any, confidence *float64, raw *string, verification string, notes *string) error {
		payload, err := json.Marshal(value)
		if err != nil {
			return err
		}
		var sourceID *uuid.UUID
		if id, ok := sourceIDs[sourceKey]; ok {
			sourceID = &id
		}
		persistence.Findings = append(persistence.Findings, application.ResearchFinding{ID: uuid.New(), ResearchRunID: run.ID, ApplicationID: task.TargetApplicationID, SourceID: sourceID, Category: category, Field: field, Value: profile.JSON(payload), Confidence: confidence, VerificationStatus: verificationOrDefault(verification), ReviewStatus: "pending", RawText: raw, Notes: notes})
		return nil
	}
	for _, finding := range result.Findings {
		if err := addFinding(finding.SourceKey, finding.Category, finding.Field, finding.Value, finding.Confidence, finding.RawText, finding.VerificationStatus, finding.Notes); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.Requirements {
		if err := addFinding(candidate.SourceKey, "requirement", candidate.Data.Title, candidate.Data, candidate.Confidence, candidate.RawText, candidate.VerificationStatus, nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.Deadlines {
		if err := addFinding(candidate.SourceKey, "deadline", candidate.Data.Title, candidate.Data, candidate.Confidence, candidate.RawText, candidate.VerificationStatus, nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.Funding {
		if err := addFinding(candidate.SourceKey, "funding", candidate.Data.FundingType, candidate.Data, candidate.Confidence, candidate.RawText, candidate.VerificationStatus, nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.Contacts {
		if err := addFinding(candidate.SourceKey, "contact", "contact", candidate.Data, candidate.Confidence, candidate.RawText, candidate.VerificationStatus, nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.Supervisors {
		if err := addFinding(candidate.SourceKey, "supervisor", candidate.Data.Name, candidate.Data, candidate.Confidence, candidate.RawText, candidate.VerificationStatus, nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.URLs {
		if err := addFinding(candidate.SourceKey, "url", candidate.Data.URLType, candidate.Data, candidate.Confidence, candidate.RawText, candidate.VerificationStatus, nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.Questionnaires {
		if err := addFinding("", "questionnaire", candidate.StableKey, candidate, nil, nil, "supported", nil); err != nil {
			return persistence, err
		}
	}
	for _, candidate := range result.ApplicationProposals {
		proposal := ApplicationProposal{Base: Base{ID: uuid.New()}, UserID: task.UserID, ResearchTaskID: task.ID, ResearchRunID: &run.ID, InstitutionID: candidate.InstitutionID, ProgrammeID: candidate.ProgrammeID, ScholarshipID: candidate.ScholarshipID, Name: candidate.Name, Country: candidate.Country, Intake: candidate.Intake, IntakeYear: candidate.IntakeYear, Summary: candidate.Summary, Status: "pending", Confidence: candidate.Confidence, ReasoningSummary: candidate.ReasoningSummary}
		if candidate.ProposedInstitution != nil {
			payload, _ := json.Marshal(candidate.ProposedInstitution)
			proposal.ProposedInstitution = profile.JSON(payload)
		}
		if candidate.ProposedProgramme != nil {
			payload, _ := json.Marshal(candidate.ProposedProgramme)
			proposal.ProposedProgramme = profile.JSON(payload)
		}
		if candidate.ProposedScholarship != nil {
			payload, _ := json.Marshal(candidate.ProposedScholarship)
			proposal.ProposedScholarship = profile.JSON(payload)
		}
		persistence.Proposals = append(persistence.Proposals, proposal)
		persistence.Outputs = append(persistence.Outputs, ResearchTaskOutput{ResearchTaskID: task.ID, OutputType: "application_proposal", EntityID: proposal.ID})
		for _, key := range candidate.SourceKeys {
			if sourceID, ok := sourceIDs[key]; ok {
				persistence.ProposalSources = append(persistence.ProposalSources, ApplicationProposalSource{ApplicationProposalID: proposal.ID, ResearchSourceID: sourceID, SourceRole: stringPointer("official")})
			}
		}
	}
	persistence.Outputs = append(persistence.Outputs, ResearchTaskOutput{ResearchTaskID: task.ID, OutputType: "research_run", EntityID: run.ID})
	for _, candidate := range result.SuggestedTasks {
		persistence.FollowUps = append(persistence.FollowUps, ResearchTask{UserID: task.UserID, ParentTaskID: &task.ID, TargetApplicationID: task.TargetApplicationID, Title: candidate.Title, Description: candidate.Description, Instructions: candidate.Instructions, TaskType: candidate.TaskType, Status: "draft", Priority: candidate.Priority})
	}
	persistence.Activities = []AgentActivity{{ResearchTaskID: &task.ID, ActivityType: "research_started", Summary: "Research provider inspected the task and supplied typed results."}, {ResearchTaskID: &task.ID, ActivityType: "proposal_created", Summary: fmt.Sprintf("Research created %d human-reviewable application proposal(s).", len(persistence.Proposals))}}
	return persistence, nil
}

func verificationOrDefault(value string) string {
	if value == "" {
		return "unverified"
	}
	return value
}

func (s *Service) ListTaskLinks(ctx context.Context, taskID uuid.UUID) ([]ResearchTaskLink, error) {
	if _, err := s.repo.GetResearchTask(ctx, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListTaskLinks(ctx, taskID)
}
func (s *Service) CreateTaskLink(ctx context.Context, taskID uuid.UUID, request TaskLinkRequest) (*ResearchTaskLink, error) {
	if _, err := s.repo.GetResearchTask(ctx, taskID); err != nil {
		return nil, err
	}
	link := &ResearchTaskLink{ResearchTaskID: taskID, Label: request.Label, URL: request.URL, LinkType: request.LinkType}
	return link, s.repo.CreateTaskLink(ctx, link)
}
func (s *Service) UpdateTaskLink(ctx context.Context, taskID, linkID uuid.UUID, request TaskLinkRequest) (*ResearchTaskLink, error) {
	link, err := s.repo.GetTaskLink(ctx, taskID, linkID)
	if err != nil {
		return nil, err
	}
	link.Label, link.URL, link.LinkType = request.Label, request.URL, request.LinkType
	return link, s.repo.UpdateTaskLink(ctx, link)
}
func (s *Service) DeleteTaskLink(ctx context.Context, taskID, linkID uuid.UUID) error {
	link, err := s.repo.GetTaskLink(ctx, taskID, linkID)
	if err != nil {
		return err
	}
	return s.repo.DeleteTaskLink(ctx, link)
}
func (s *Service) ListTaskOutputs(ctx context.Context, taskID uuid.UUID) ([]ResearchTaskOutput, error) {
	if _, err := s.repo.GetResearchTask(ctx, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListTaskOutputs(ctx, taskID)
}
func (s *Service) ListTaskRuns(ctx context.Context, taskID uuid.UUID) ([]application.ResearchRun, error) {
	if _, err := s.repo.GetResearchTask(ctx, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListTaskRuns(ctx, taskID)
}
func (s *Service) ListRunSources(ctx context.Context, taskID, runID uuid.UUID) ([]application.ResearchSource, error) {
	if err := s.requireTaskRun(ctx, taskID, runID); err != nil {
		return nil, err
	}
	return s.repo.ListRunSources(ctx, runID)
}
func (s *Service) ListRunFindings(ctx context.Context, taskID, runID uuid.UUID) ([]application.ResearchFinding, error) {
	if err := s.requireTaskRun(ctx, taskID, runID); err != nil {
		return nil, err
	}
	return s.repo.ListRunFindings(ctx, runID)
}

func (s *Service) CreateAgentResearchSource(ctx context.Context, taskID uuid.UUID, request AgentResearchSourceRequest) (*application.ResearchSource, error) {
	run, err := s.taskRun(ctx, taskID, request.ResearchRunID)
	if err != nil {
		return nil, err
	}
	retrievedAt := s.now()
	if request.RetrievedAt != nil {
		retrievedAt = request.RetrievedAt.UTC()
	}
	value := &application.ResearchSource{ID: uuid.New(), ResearchRunID: run.ID, ApplicationID: run.ApplicationID, URL: request.URL, Title: request.Title, Publisher: request.Publisher, SourceType: request.SourceType, IsOfficial: request.IsOfficial, RetrievedAt: retrievedAt, PublishedAt: request.PublishedAt, ContentHash: request.ContentHash, Notes: request.Notes}
	activity := &AgentActivity{ID: uuid.New(), ResearchTaskID: &taskID, ApplicationID: run.ApplicationID, ActivityType: "source_added", Summary: "Authenticated research agent added a research source."}
	if err := s.repo.CreateAgentSource(ctx, value, activity); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Service) CreateAgentResearchFinding(ctx context.Context, taskID uuid.UUID, request AgentResearchFindingRequest) (*application.ResearchFinding, error) {
	run, err := s.taskRun(ctx, taskID, request.ResearchRunID)
	if err != nil {
		return nil, err
	}
	if request.SourceID != nil {
		sources, sourceErr := s.repo.ListRunSources(ctx, run.ID)
		if sourceErr != nil {
			return nil, sourceErr
		}
		matched := false
		for _, source := range sources {
			if source.ID == *request.SourceID {
				matched = true
				break
			}
		}
		if !matched {
			return nil, validationError("sourceId does not belong to the research run")
		}
	}
	value := &application.ResearchFinding{ID: uuid.New(), ResearchRunID: run.ID, ApplicationID: run.ApplicationID, SourceID: request.SourceID, Category: request.Category, Field: request.Field, Value: profile.JSON(request.Value), Confidence: request.Confidence, VerificationStatus: request.VerificationStatus, ReviewStatus: "pending", RawText: request.RawText, Notes: request.Notes}
	activity := &AgentActivity{ID: uuid.New(), ResearchTaskID: &taskID, ApplicationID: run.ApplicationID, ActivityType: "finding_added", Summary: "Authenticated research agent added a pending research finding."}
	if err := s.repo.CreateAgentFinding(ctx, value, activity); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Service) CreateAgentApplicationProposal(ctx context.Context, taskID uuid.UUID, request AgentApplicationProposalRequest) (*ApplicationProposalResponse, error) {
	task, err := s.GetResearchTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	run, err := s.taskRun(ctx, taskID, request.ResearchRunID)
	if err != nil {
		return nil, err
	}
	value := &ApplicationProposal{Base: Base{ID: uuid.New()}, UserID: task.UserID, ResearchTaskID: task.ID, ResearchRunID: &run.ID, InstitutionID: request.InstitutionID, ProgrammeID: request.ProgrammeID, ScholarshipID: request.ScholarshipID, ProposedInstitution: profile.JSON(request.ProposedInstitution), ProposedProgramme: profile.JSON(request.ProposedProgramme), ProposedScholarship: profile.JSON(request.ProposedScholarship), Name: strings.TrimSpace(request.Name), Country: request.Country, Intake: request.Intake, IntakeYear: request.IntakeYear, Summary: request.Summary, Status: "pending", Confidence: request.Confidence, ReasoningSummary: request.ReasoningSummary}
	sources, err := s.repo.ListRunSources(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	allowed := map[uuid.UUID]application.ResearchSource{}
	for _, source := range sources {
		allowed[source.ID] = source
	}
	links := make([]ApplicationProposalSource, 0, len(request.SourceIDs))
	for _, sourceID := range request.SourceIDs {
		if _, ok := allowed[sourceID]; !ok {
			return nil, validationError("sourceIds must belong to the research run")
		}
		role := "supporting"
		links = append(links, ApplicationProposalSource{ID: uuid.New(), ApplicationProposalID: value.ID, ResearchSourceID: sourceID, SourceRole: &role})
	}
	output := &ResearchTaskOutput{ID: uuid.New(), ResearchTaskID: task.ID, OutputType: "application_proposal", EntityID: value.ID}
	activity := &AgentActivity{ID: uuid.New(), ResearchTaskID: &task.ID, ActivityType: "proposal_created", Summary: "Authenticated research agent created an application proposal for human review."}
	task.Status, task.CompletedAt = "review_required", nil
	if err := s.repo.CreateAgentProposal(ctx, task, value, links, output, activity); err != nil {
		return nil, err
	}
	records := make([]ProposalSourceRecord, 0, len(links))
	for _, link := range links {
		records = append(records, ProposalSourceRecord{Link: link, Source: allowed[link.ResearchSourceID]})
	}
	response := proposalResponse(value, records)
	return &response, nil
}

func (s *Service) taskRun(ctx context.Context, taskID, runID uuid.UUID) (*application.ResearchRun, error) {
	if _, err := s.GetResearchTask(ctx, taskID); err != nil {
		return nil, err
	}
	runs, err := s.repo.ListTaskRuns(ctx, taskID)
	if err != nil {
		return nil, err
	}
	for i := range runs {
		if runs[i].ID == runID {
			return &runs[i], nil
		}
	}
	return nil, application.ErrResearchRunNotFound
}

func (s *Service) GetApplicationEffectiveProfile(ctx context.Context, applicationID uuid.UUID) (*profile.EffectiveProfileResponse, error) {
	value, err := s.applications.Get(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	return s.profiles.ResolveEffectiveProfile(ctx, value.ApplicantProfileID)
}
func (s *Service) requireTaskRun(ctx context.Context, taskID, runID uuid.UUID) error {
	runs, err := s.repo.ListTaskRuns(ctx, taskID)
	if err != nil {
		return err
	}
	for _, run := range runs {
		if run.ID == runID {
			return nil
		}
	}
	return application.ErrResearchRunNotFound
}

func (s *Service) ListProposals(ctx context.Context, userID *uuid.UUID, status *string) ([]ApplicationProposal, error) {
	if actor, ok := principal.PrincipalFromContext(ctx); ok && !actor.IsSystem() {
		userID = actor.UserID
	}
	return s.repo.ListProposals(ctx, userID, status)
}
func (s *Service) ListProposalResponses(ctx context.Context, userID *uuid.UUID, status *string) ([]ApplicationProposalResponse, error) {
	if actor, ok := principal.PrincipalFromContext(ctx); ok && !actor.IsSystem() {
		userID = actor.UserID
	}
	items, err := s.repo.ListProposals(ctx, userID, status)
	if err != nil {
		return nil, err
	}
	result := make([]ApplicationProposalResponse, 0, len(items))
	for i := range items {
		sources, sourceErr := s.repo.ListProposalSources(ctx, items[i].ID)
		if sourceErr != nil {
			return nil, sourceErr
		}
		result = append(result, proposalResponse(&items[i], sources))
	}
	return result, nil
}
func (s *Service) GetProposal(ctx context.Context, id uuid.UUID) (*ApplicationProposalResponse, error) {
	proposal, err := s.repo.GetProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, proposal.UserID) != nil {
		return nil, ErrApplicationProposalNotFound
	}
	sources, err := s.repo.ListProposalSources(ctx, id)
	if err != nil {
		return nil, err
	}
	response := proposalResponse(proposal, sources)
	return &response, nil
}

func (s *Service) ApproveProposal(ctx context.Context, id uuid.UUID, request ApproveProposalRequest) (*application.Application, error) {
	proposal, err := s.repo.GetProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, proposal.UserID) != nil {
		return nil, ErrApplicationProposalNotFound
	}
	if proposal.Status != "pending" {
		return nil, ErrInvalidProposalState
	}
	parent, err := s.profiles.Get(ctx, request.ParentProfileID)
	if err != nil {
		return nil, err
	}
	if parent.UserID != proposal.UserID || (parent.ProfileType != profile.ProfileTypeMaster && parent.ProfileType != profile.ProfileTypeDomain) {
		return nil, ErrProposalParentRequired
	}
	task, err := s.repo.GetResearchTask(ctx, proposal.ResearchTaskID)
	if err != nil {
		return nil, err
	}
	created, err := s.repo.ApproveProposal(ctx, proposal, task, parent)
	if err == nil {
		appLogger.Info(ctx, "application proposal approved", "proposal_id", id, "application_id", created.ID)
	}
	return created, err
}

func (s *Service) ReviewProposal(ctx context.Context, id uuid.UUID, action string) (*ApplicationProposal, error) {
	proposal, err := s.repo.GetProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	task, err := s.repo.GetResearchTask(ctx, proposal.ResearchTaskID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	switch action {
	case "reject":
		if proposal.Status != "pending" {
			return nil, ErrInvalidProposalState
		}
		proposal.Status, proposal.ReviewedAt = "rejected", &now
		others, listErr := s.repo.ListProposals(ctx, &proposal.UserID, stringPointer("pending"))
		if listErr != nil {
			return nil, listErr
		}
		remaining := false
		for _, other := range others {
			if other.ResearchTaskID == task.ID && other.ID != proposal.ID {
				remaining = true
				break
			}
		}
		if remaining {
			task.Status, task.CompletedAt = "review_required", nil
		} else {
			task.Status, task.CompletedAt = "completed", &now
		}
	case "reopen":
		if proposal.Status != "rejected" {
			return nil, ErrInvalidProposalState
		}
		proposal.Status, proposal.ReviewedAt = "pending", nil
		task.Status, task.CompletedAt = "review_required", nil
	default:
		return nil, ErrInvalidProposalState
	}
	activity := AgentActivity{ResearchTaskID: &task.ID, ActivityType: "proposal_" + action, Summary: "Human review changed the application proposal state to " + proposal.Status + "."}
	return proposal, s.repo.UpdateProposal(ctx, proposal, task, activity)
}

func (s *Service) ListFields(ctx context.Context, appID uuid.UUID) ([]ApplicationField, error) {
	if _, err := s.applications.Get(ctx, appID); err != nil {
		return nil, err
	}
	return s.repo.ListFields(ctx, appID)
}
func (s *Service) CreateField(ctx context.Context, appID uuid.UUID, request ApplicationFieldRequest) (*ApplicationField, error) {
	if _, err := s.applications.Get(ctx, appID); err != nil {
		return nil, err
	}
	status := "empty"
	if request.Status != nil {
		status = *request.Status
	}
	item := &ApplicationField{ApplicationID: appID, Key: request.Key, Label: request.Label, Value: profile.JSON(request.Value), ValueType: request.ValueType, Status: status, SourceType: request.SourceType, SourceEntityID: request.SourceEntityID, Notes: request.Notes}
	return item, s.repo.CreateField(ctx, item)
}
func (s *Service) UpdateField(ctx context.Context, appID, id uuid.UUID, request ApplicationFieldRequest) (*ApplicationField, error) {
	item, err := s.repo.GetField(ctx, appID, id)
	if err != nil {
		return nil, err
	}
	item.Key, item.Label, item.Value, item.ValueType, item.SourceType, item.SourceEntityID, item.Notes = request.Key, request.Label, profile.JSON(request.Value), request.ValueType, request.SourceType, request.SourceEntityID, request.Notes
	if request.Status != nil {
		item.Status = *request.Status
	}
	return item, s.repo.UpdateField(ctx, item)
}
func (s *Service) DeleteField(ctx context.Context, appID, id uuid.UUID) error {
	item, err := s.repo.GetField(ctx, appID, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteField(ctx, item)
}

func (s *Service) ListQuestionnaires(ctx context.Context, appID uuid.UUID) ([]ApplicationQuestionnaire, error) {
	if _, err := s.applications.Get(ctx, appID); err != nil {
		return nil, err
	}
	return s.repo.ListQuestionnaires(ctx, appID)
}
func (s *Service) GetQuestionnaire(ctx context.Context, appID, id uuid.UUID) (*QuestionnaireResponse, error) {
	item, err := s.repo.GetQuestionnaire(ctx, appID, id)
	if err != nil {
		return nil, err
	}
	questions, err := s.repo.ListQuestions(ctx, id)
	if err != nil {
		return nil, err
	}
	response := questionnaireResponse(item, questions)
	return &response, nil
}
func (s *Service) CreateQuestionnaire(ctx context.Context, appID uuid.UUID, request QuestionnaireRequest) (*ApplicationQuestionnaire, error) {
	if _, err := s.applications.Get(ctx, appID); err != nil {
		return nil, err
	}
	status := "not_started"
	if request.Status != nil {
		status = *request.Status
	}
	item := &ApplicationQuestionnaire{ApplicationID: appID, Title: request.Title, Description: request.Description, QuestionnaireType: request.QuestionnaireType, Status: status, SourceURL: request.SourceURL}
	return item, s.repo.CreateQuestionnaire(ctx, item)
}
func (s *Service) UpdateQuestionnaire(ctx context.Context, appID, id uuid.UUID, request QuestionnaireRequest) (*ApplicationQuestionnaire, error) {
	item, err := s.repo.GetQuestionnaire(ctx, appID, id)
	if err != nil {
		return nil, err
	}
	item.Title, item.Description, item.QuestionnaireType, item.SourceURL = request.Title, request.Description, request.QuestionnaireType, request.SourceURL
	if request.Status != nil {
		item.Status = *request.Status
	}
	return item, s.repo.UpdateQuestionnaire(ctx, item)
}
func (s *Service) DeleteQuestionnaire(ctx context.Context, appID, id uuid.UUID) error {
	item, err := s.repo.GetQuestionnaire(ctx, appID, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteQuestionnaire(ctx, item)
}

func (s *Service) ListQuestions(ctx context.Context, questionnaireID uuid.UUID) ([]ApplicationQuestion, error) {
	return s.repo.ListQuestions(ctx, questionnaireID)
}
func (s *Service) CreateQuestion(ctx context.Context, questionnaireID uuid.UUID, request QuestionRequest) (*ApplicationQuestion, error) {
	status := "unanswered"
	if request.Status != nil {
		status = *request.Status
	}
	item := &ApplicationQuestion{QuestionnaireID: questionnaireID, Key: request.Key, Prompt: request.Prompt, HelpText: request.HelpText, QuestionType: request.QuestionType, IsRequired: request.IsRequired, WordLimit: request.WordLimit, CharacterLimit: request.CharacterLimit, SortOrder: request.SortOrder, Options: profile.JSON(request.Options), Status: status}
	return item, s.repo.CreateQuestion(ctx, item)
}
func (s *Service) UpdateQuestion(ctx context.Context, questionnaireID, id uuid.UUID, request QuestionRequest) (*ApplicationQuestion, error) {
	item, err := s.repo.GetQuestion(ctx, questionnaireID, id)
	if err != nil {
		return nil, err
	}
	item.Key, item.Prompt, item.HelpText, item.QuestionType, item.IsRequired, item.WordLimit, item.CharacterLimit, item.SortOrder, item.Options = request.Key, request.Prompt, request.HelpText, request.QuestionType, request.IsRequired, request.WordLimit, request.CharacterLimit, request.SortOrder, profile.JSON(request.Options)
	if request.Status != nil {
		item.Status = *request.Status
	}
	return item, s.repo.UpdateQuestion(ctx, item)
}
func (s *Service) DeleteQuestion(ctx context.Context, questionnaireID, id uuid.UUID) error {
	item, err := s.repo.GetQuestion(ctx, questionnaireID, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteQuestion(ctx, item)
}

func (s *Service) ListAnswers(ctx context.Context, questionID uuid.UUID) ([]ApplicationAnswer, error) {
	return s.repo.ListAnswers(ctx, questionID)
}
func (s *Service) CreateAnswer(ctx context.Context, questionID uuid.UUID, request AnswerRequest) (*ApplicationAnswer, error) {
	question, err := s.findQuestion(ctx, questionID)
	if err != nil {
		return nil, err
	}
	status := "draft"
	if request.Status != nil {
		status = *request.Status
	}
	item := &ApplicationAnswer{QuestionID: questionID, Value: profile.JSON(request.Value), DraftText: request.DraftText, Status: status, AnswerSource: request.AnswerSource, SourceEntityType: request.SourceEntityType, SourceEntityID: request.SourceEntityID, Confidence: request.Confidence, CreatedBy: request.CreatedBy}
	if err = s.repo.CreateAnswer(ctx, item); err != nil {
		return nil, err
	}
	question.Status = questionStatusForAnswer(item.Status)
	if err = s.repo.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}
	_ = s.repo.RefreshQuestionnaireStatus(ctx, question.QuestionnaireID)
	return item, nil
}
func (s *Service) UpdateAnswer(ctx context.Context, questionID, id uuid.UUID, request AnswerRequest) (*ApplicationAnswer, error) {
	question, err := s.findQuestion(ctx, questionID)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.GetAnswer(ctx, questionID, id)
	if err != nil {
		return nil, err
	}
	item.Value, item.DraftText, item.AnswerSource, item.SourceEntityType, item.SourceEntityID, item.Confidence, item.CreatedBy = profile.JSON(request.Value), request.DraftText, request.AnswerSource, request.SourceEntityType, request.SourceEntityID, request.Confidence, request.CreatedBy
	if request.Status != nil {
		item.Status = *request.Status
	}
	if err = s.repo.UpdateAnswer(ctx, item); err != nil {
		return nil, err
	}
	question.Status = questionStatusForAnswer(item.Status)
	if err = s.repo.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}
	_ = s.repo.RefreshQuestionnaireStatus(ctx, question.QuestionnaireID)
	return item, nil
}
func (s *Service) ReviewAnswer(ctx context.Context, questionID, id uuid.UUID, action string) (*ApplicationAnswer, error) {
	question, err := s.findQuestion(ctx, questionID)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.GetAnswer(ctx, questionID, id)
	if err != nil {
		return nil, err
	}
	if action == "approve" {
		item.Status = "approved"
		question.Status = "approved"
	} else if action == "reject" {
		item.Status = "rejected"
		question.Status = "needs_review"
	} else {
		return nil, validationError("invalid answer review action")
	}
	if err = s.repo.UpdateAnswer(ctx, item); err != nil {
		return nil, err
	}
	if err = s.repo.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}
	_ = s.repo.RefreshQuestionnaireStatus(ctx, question.QuestionnaireID)
	return item, nil
}
func (s *Service) findQuestion(ctx context.Context, id uuid.UUID) (*ApplicationQuestion, error) {
	return s.repo.GetQuestionByID(ctx, id)
}
func (s *Service) RegenerateQuestion(ctx context.Context, questionID uuid.UUID) (*ApplicationPrefillRun, error) {
	question, err := s.repo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	question.Status = "unanswered"
	if err = s.repo.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}
	questionnaire, err := s.repo.GetQuestionnaireByID(ctx, question.QuestionnaireID)
	if err != nil {
		return nil, err
	}
	return s.RunPrefill(ctx, questionnaire.ApplicationID, PrefillRequest{Trigger: "regeneration", Regenerate: true})
}

func (s *Service) ListInformationRequests(ctx context.Context, filters InformationRequestFilters) ([]InformationRequest, error) {
	if actor, ok := principal.PrincipalFromContext(ctx); ok && !actor.IsSystem() {
		filters.UserID = actor.UserID
	}
	return s.repo.ListInformationRequests(ctx, filters)
}
func (s *Service) GetInformationRequest(ctx context.Context, id uuid.UUID) (*InformationRequestResponseDTO, error) {
	item, err := s.repo.GetInformationRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.EnforceOwner(ctx, item.UserID) != nil {
		return nil, ErrInformationRequestNotFound
	}
	responses, err := s.repo.ListInformationResponses(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := informationRequestResponse(item, responses)
	return &dto, nil
}
func (s *Service) CreateInformationRequest(ctx context.Context, request CreateInformationRequest) (*InformationRequest, error) {
	if principal.EnforceOwner(ctx, request.UserID) != nil {
		return nil, ErrInformationRequestNotFound
	}
	_, authenticated := principal.PrincipalFromContext(ctx)
	if authenticated && request.ApplicationID != nil {
		if _, err := s.applications.Get(ctx, *request.ApplicationID); err != nil {
			return nil, err
		}
	}
	if authenticated && request.ResearchTaskID != nil {
		if _, err := s.GetResearchTask(ctx, *request.ResearchTaskID); err != nil {
			return nil, err
		}
	}
	if authenticated && request.QuestionID != nil {
		question, err := s.repo.GetQuestionByID(ctx, *request.QuestionID)
		if err != nil {
			return nil, err
		}
		questionnaire, err := s.repo.GetQuestionnaireByID(ctx, question.QuestionnaireID)
		if err != nil {
			return nil, err
		}
		if _, err := s.applications.Get(ctx, questionnaire.ApplicationID); err != nil {
			return nil, err
		}
	}
	if authenticated && request.ApplicationFieldID != nil {
		if request.ApplicationID == nil {
			return nil, validationError("applicationId is required for an application field request")
		}
		if _, err := s.repo.GetField(ctx, *request.ApplicationID, *request.ApplicationFieldID); err != nil {
			return nil, err
		}
	}
	item := InformationRequest{UserID: request.UserID, ApplicationID: request.ApplicationID, ResearchTaskID: request.ResearchTaskID, QuestionnaireID: request.QuestionnaireID, QuestionID: request.QuestionID, ApplicationFieldID: request.ApplicationFieldID, RequestType: request.RequestType, Title: request.Title, Prompt: request.Prompt, Context: request.Context, Status: "pending", Priority: request.Priority, ResponseType: request.ResponseType, Options: profile.JSON(request.Options), CreatedBy: request.CreatedBy}
	if item.QuestionID == nil && item.ApplicationFieldID == nil {
		return nil, ErrInformationTargetRequired
	}
	existing, err := s.repo.FindOpenInformationRequest(ctx, item)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrInformationRequestNotFound) {
		return nil, err
	}
	return &item, s.repo.CreateInformationRequest(ctx, &item)
}
func (s *Service) RespondInformationRequest(ctx context.Context, id uuid.UUID, request RespondInformationRequest) (*InformationRequest, error) {
	item, err := s.repo.GetInformationRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.Status != "pending" && item.Status != "reopened" && item.Status != "answered" {
		return nil, ErrInvalidInformationState
	}
	if len(request.ResponseValue) == 0 && (request.ResponseText == nil || strings.TrimSpace(*request.ResponseText) == "") {
		return nil, validationError("responseValue or responseText is required")
	}
	item.ResponseValue, item.ResponseText, item.Status = profile.JSON(request.ResponseValue), request.ResponseText, "answered"
	response := &InformationRequestResponse{InformationRequestID: item.ID, ResponseValue: item.ResponseValue, ResponseText: item.ResponseText, SubmittedBy: request.SubmittedBy}
	return item, s.repo.RespondInformationRequest(ctx, item, response)
}
func (s *Service) ProcessInformationRequest(ctx context.Context, id uuid.UUID) (*InformationRequest, error) {
	item, err := s.repo.GetInformationRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.Status != "answered" {
		return nil, ErrInvalidInformationState
	}
	if err = s.repo.ProcessInformationRequest(ctx, item); err != nil {
		return nil, err
	}
	if item.QuestionID != nil {
		if question, questionErr := s.repo.GetQuestionByID(ctx, *item.QuestionID); questionErr == nil {
			_ = s.repo.RefreshQuestionnaireStatus(ctx, question.QuestionnaireID)
		}
	}
	return s.repo.GetInformationRequest(ctx, id)
}
func (s *Service) TransitionInformationRequest(ctx context.Context, id uuid.UUID, action string) (*InformationRequest, error) {
	item, err := s.repo.GetInformationRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	now := s.now()
	switch action {
	case "reopen":
		if item.Status != "completed" && item.Status != "cancelled" {
			return nil, ErrInvalidInformationState
		}
		item.Status, item.ReopenedAt, item.CompletedAt = "reopened", &now, nil
	case "cancel":
		if item.Status == "completed" {
			return nil, ErrInvalidInformationState
		}
		item.Status = "cancelled"
	default:
		return nil, ErrInvalidInformationState
	}
	return item, s.repo.UpdateInformationRequest(ctx, item)
}

func (s *Service) RunPrefill(ctx context.Context, applicationID uuid.UUID, request PrefillRequest) (*ApplicationPrefillRun, error) {
	data, err := s.repo.LoadPrefillInput(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	effective, err := s.profiles.ResolveEffectiveProfile(ctx, data.Application.ApplicantProfileID)
	if err != nil {
		return nil, err
	}
	input := PrefillInput{Application: applicationResponse(data.Application), EffectiveProfile: effective, Requirements: data.Requirements, Fields: data.Fields, Questionnaires: data.Questionnaires, CompletedInformationRequests: data.CompletedInfo, ResearchFindings: data.Findings, Regenerate: request.Regenerate}
	result, err := s.prefiller.Prefill(ctx, input)
	if err != nil {
		return nil, err
	}
	trigger := request.Trigger
	if trigger == "" {
		trigger = "manual"
	}
	now, provider, model := s.now(), s.prefiller.ProviderName(), s.prefiller.ModelName()
	run := &ApplicationPrefillRun{Base: Base{ID: uuid.New()}, ApplicationID: applicationID, Status: "running", Trigger: trigger, AgentProvider: &provider, AgentModel: &model, StartedAt: &now}
	persistence := PrefillPersistence{Run: run, Activities: []AgentActivity{{ApplicationID: &applicationID, PrefillRunID: &run.ID, ActivityType: "prefill_started", Summary: "Controlled application prefill evaluated supported facts and missing information."}}}
	completedTargets := completedInformationTargets(data.CompletedInfo)
	for _, candidate := range result.FieldUpdates {
		payload, marshalErr := json.Marshal(candidate.Value)
		if marshalErr != nil {
			return nil, marshalErr
		}
		field := ApplicationField{Base: Base{ID: uuid.New()}, ApplicationID: applicationID, Key: candidate.Key, Label: candidate.Label, Value: profile.JSON(payload), ValueType: candidate.ValueType, Status: candidate.Status, SourceType: &candidate.SourceType, SourceEntityID: candidate.SourceEntityID, PrefillRunID: &run.ID}
		persistence.Fields = append(persistence.Fields, field)
		persistence.Actions = append(persistence.Actions, prefillAction(run.ID, "field_filled", "application_field", &field.ID, "Supported application field filled."))
	}
	for _, candidate := range result.QuestionAnswers {
		payload, marshalErr := json.Marshal(candidate.Value)
		if marshalErr != nil {
			return nil, marshalErr
		}
		answerSource, createdBy := candidate.AnswerSource, "agent"
		answer := ApplicationAnswer{Base: Base{ID: uuid.New()}, QuestionID: candidate.QuestionID, Value: profile.JSON(payload), DraftText: candidate.DraftText, Status: candidate.Status, AnswerSource: &answerSource, SourceEntityType: candidate.SourceEntityType, SourceEntityID: candidate.SourceEntityID, Confidence: candidate.Confidence, CreatedBy: &createdBy, PrefillRunID: &run.ID, AgentProvider: &provider, AgentModel: &model}
		persistence.Answers = append(persistence.Answers, answer)
		action := "answer_generated"
		persistence.Actions = append(persistence.Actions, prefillAction(run.ID, action, "application_answer", &answer.ID, "Supported answer generated with provenance."))
	}
	for _, candidate := range result.InformationRequests {
		if candidate.QuestionID != nil && completedTargets["question:"+candidate.QuestionID.String()] {
			continue
		}
		if candidate.ApplicationFieldID != nil && completedTargets["field:"+candidate.ApplicationFieldID.String()] {
			continue
		}
		item := InformationRequest{Base: Base{ID: uuid.New()}, UserID: data.Application.UserID, ApplicationID: &applicationID, QuestionnaireID: candidate.QuestionnaireID, QuestionID: candidate.QuestionID, ApplicationFieldID: candidate.ApplicationFieldID, RequestType: candidate.RequestType, Title: candidate.Title, Prompt: candidate.Prompt, Context: candidate.Context, Status: "pending", Priority: candidate.Priority, ResponseType: candidate.ResponseType, CreatedBy: "agent"}
		persistence.InformationRequests = append(persistence.InformationRequests, item)
		persistence.Actions = append(persistence.Actions, prefillAction(run.ID, "information_requested", "information_request", &item.ID, "Missing information requested instead of guessed."))
	}
	for _, task := range result.SuggestedTasks {
		task.ApplicationID, task.Source = applicationID, stringPointer("prefill")
		persistence.Tasks = append(persistence.Tasks, task)
	}
	if len(persistence.InformationRequests) == 0 {
		persistence.Activities = append(persistence.Activities, AgentActivity{ApplicationID: &applicationID, PrefillRunID: &run.ID, ActivityType: "prefill_completed", Summary: "Prefill completed without creating duplicate information requests."})
	}
	if err = s.repo.PersistPrefillResult(ctx, persistence); err != nil {
		return nil, err
	}
	for _, bundle := range data.Questionnaires {
		_ = s.repo.RefreshQuestionnaireStatus(ctx, bundle.Questionnaire.ID)
	}
	return run, nil
}

func completedInformationTargets(items []InformationRequest) map[string]bool {
	result := map[string]bool{}
	for _, item := range items {
		if item.QuestionID != nil {
			result["question:"+item.QuestionID.String()] = true
		}
		if item.ApplicationFieldID != nil {
			result["field:"+item.ApplicationFieldID.String()] = true
		}
	}
	return result
}
func prefillAction(runID uuid.UUID, actionType, targetType string, targetID *uuid.UUID, summary string) PrefillAction {
	return PrefillAction{PrefillRunID: runID, ActionType: actionType, TargetType: targetType, TargetID: targetID, Status: "complete", Summary: &summary}
}
func (s *Service) PreparationSummary(ctx context.Context, applicationID uuid.UUID) (PreparationSummary, error) {
	if _, err := s.applications.Get(ctx, applicationID); err != nil {
		return PreparationSummary{}, err
	}
	return s.repo.PreparationSummary(ctx, applicationID)
}
func (s *Service) ApplicationPreparation(ctx context.Context, applicationID uuid.UUID) (application.PreparationComponents, error) {
	summary, err := s.repo.PreparationSummary(ctx, applicationID)
	if err != nil {
		return application.PreparationComponents{}, err
	}
	return application.PreparationComponents{FieldsCompleted: summary.FieldsCompleted, FieldsTotal: summary.FieldsTotal, QuestionnairesCompleted: summary.QuestionnairesCompleted, QuestionnairesTotal: summary.QuestionnairesTotal, PendingInformation: summary.PendingInformation, RequiredQuestionsAnswered: summary.RequiredQuestionsAnswered, RequiredQuestionsTotal: summary.RequiredQuestionsTotal}, nil
}
