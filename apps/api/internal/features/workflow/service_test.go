package workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/scholarship-os/apps/api/internal/features/application"
	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	"github.com/example/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	tasks       map[uuid.UUID]*ResearchTask
	links       map[uuid.UUID][]ResearchTaskLink
	persistence *ResearchPersistence
	proposal    *ApplicationProposal
	approved    bool
	info        *InformationRequest
	history     []InformationRequestResponse
	prefillData *PrefillData
	prefill     *PrefillPersistence
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{tasks: map[uuid.UUID]*ResearchTask{}, links: map[uuid.UUID][]ResearchTaskLink{}}
}
func (f *fakeRepository) CreateResearchTask(_ context.Context, task *ResearchTask, links []ResearchTaskLink) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	f.tasks[task.ID] = task
	for i := range links {
		links[i].ID, links[i].ResearchTaskID = uuid.New(), task.ID
	}
	f.links[task.ID] = links
	return nil
}
func (f *fakeRepository) GetResearchTask(_ context.Context, id uuid.UUID) (*ResearchTask, error) {
	task := f.tasks[id]
	if task == nil {
		return nil, ErrResearchTaskNotFound
	}
	return task, nil
}
func (f *fakeRepository) UpdateResearchTask(_ context.Context, task *ResearchTask) error {
	f.tasks[task.ID] = task
	return nil
}
func (f *fakeRepository) ListTaskLinks(_ context.Context, id uuid.UUID) ([]ResearchTaskLink, error) {
	return f.links[id], nil
}
func (f *fakeRepository) PersistResearchTaskResult(_ context.Context, task *ResearchTask, value ResearchPersistence) error {
	f.persistence = &value
	f.tasks[task.ID] = task
	if len(value.Proposals) > 0 {
		f.proposal = &value.Proposals[0]
	}
	return nil
}
func (f *fakeRepository) FailResearchTask(context.Context, *ResearchTask, *application.ResearchRun, error) error {
	return nil
}
func (f *fakeRepository) GetProposal(_ context.Context, id uuid.UUID) (*ApplicationProposal, error) {
	if f.proposal == nil || f.proposal.ID != id {
		return nil, ErrApplicationProposalNotFound
	}
	return f.proposal, nil
}
func (f *fakeRepository) ListProposals(_ context.Context, userID *uuid.UUID, status *string) ([]ApplicationProposal, error) {
	if f.proposal == nil || (userID != nil && f.proposal.UserID != *userID) || (status != nil && f.proposal.Status != *status) {
		return []ApplicationProposal{}, nil
	}
	return []ApplicationProposal{*f.proposal}, nil
}
func (f *fakeRepository) UpdateProposal(_ context.Context, proposal *ApplicationProposal, task *ResearchTask, _ AgentActivity) error {
	f.proposal = proposal
	f.tasks[task.ID] = task
	return nil
}
func (f *fakeRepository) ApproveProposal(_ context.Context, proposal *ApplicationProposal, task *ResearchTask, parent *profile.ApplicantProfile) (*application.Application, error) {
	if f.approved {
		return nil, ErrInvalidProposalState
	}
	f.approved = true
	proposal.Status, task.Status = "approved", "completed"
	return &application.Application{Base: application.Base{ID: uuid.New()}, UserID: proposal.UserID, ApplicantProfileID: uuid.New(), Name: proposal.Name}, nil
}
func (f *fakeRepository) FindOpenInformationRequest(_ context.Context, candidate InformationRequest) (*InformationRequest, error) {
	if f.info != nil && f.info.Status != "completed" {
		return f.info, nil
	}
	return nil, ErrInformationRequestNotFound
}
func (f *fakeRepository) CreateInformationRequest(_ context.Context, item *InformationRequest) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	f.info = item
	return nil
}
func (f *fakeRepository) GetInformationRequest(_ context.Context, id uuid.UUID) (*InformationRequest, error) {
	if f.info == nil || f.info.ID != id {
		return nil, ErrInformationRequestNotFound
	}
	return f.info, nil
}
func (f *fakeRepository) RespondInformationRequest(_ context.Context, item *InformationRequest, response *InformationRequestResponse) error {
	f.info, f.history = item, append(f.history, *response)
	return nil
}
func (f *fakeRepository) ProcessInformationRequest(_ context.Context, item *InformationRequest) error {
	item.Status = "completed"
	f.info = item
	return nil
}
func (f *fakeRepository) GetQuestionByID(_ context.Context, id uuid.UUID) (*ApplicationQuestion, error) {
	return &ApplicationQuestion{Base: Base{ID: id}, QuestionnaireID: uuid.New()}, nil
}
func (f *fakeRepository) UpdateInformationRequest(_ context.Context, item *InformationRequest) error {
	f.info = item
	return nil
}
func (f *fakeRepository) LoadPrefillInput(context.Context, uuid.UUID) (*PrefillData, error) {
	return f.prefillData, nil
}
func (f *fakeRepository) PersistPrefillResult(_ context.Context, data PrefillPersistence) error {
	f.prefill = &data
	data.Run.Status = "review_required"
	return nil
}
func (f *fakeRepository) RefreshQuestionnaireStatus(context.Context, uuid.UUID) error { return nil }

type fakeUsers struct{ id uuid.UUID }

func (*fakeUsers) Create(context.Context, *user.User) error { return nil }
func (f *fakeUsers) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	if id != f.id {
		return nil, user.ErrNotFound
	}
	return &user.User{ID: id, Email: "alex@example.test"}, nil
}

type fakeApplications struct{ item *application.Application }

func (f *fakeApplications) Get(_ context.Context, id uuid.UUID) (*application.Application, error) {
	if f.item == nil || f.item.ID != id {
		return nil, application.ErrApplicationNotFound
	}
	return f.item, nil
}

type fakeProfiles struct {
	item      *profile.ApplicantProfile
	effective *profile.EffectiveProfileResponse
}

func (f *fakeProfiles) Get(_ context.Context, id uuid.UUID) (*profile.ApplicantProfile, error) {
	if f.item == nil || f.item.ID != id {
		return nil, profile.ErrProfileNotFound
	}
	return f.item, nil
}
func (f *fakeProfiles) ResolveEffectiveProfile(context.Context, uuid.UUID) (*profile.EffectiveProfileResponse, error) {
	if f.effective == nil {
		return nil, errors.New("missing effective profile")
	}
	return f.effective, nil
}

func newTestService(repo *fakeRepository, userID uuid.UUID, apps *fakeApplications, profiles *fakeProfiles) *Service {
	return NewService(repo, &fakeUsers{id: userID}, apps, profiles, NewMockWebResearchProvider(), NewMockApplicationPrefiller())
}

func TestResearchTaskLifecycleCreatesProposalWithoutApplication(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	repo := newFakeRepository()
	service := newTestService(repo, userID, &fakeApplications{}, &fakeProfiles{})
	task, err := service.CreateResearchTask(ctx, CreateResearchTaskRequest{UserID: userID, Title: "Example Scholarship", TaskType: "scholarship_research", Links: []TaskLinkRequest{{URL: "https://example.test/official"}}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != "queued" {
		t.Fatalf("status = %s", task.Status)
	}
	if _, err = service.StartResearchTask(ctx, task.ID); err != nil {
		t.Fatal(err)
	}
	if task.Status != "review_required" {
		t.Fatalf("status after research = %s", task.Status)
	}
	if repo.persistence == nil || len(repo.persistence.Proposals) != 1 || len(repo.persistence.Sources) == 0 || len(repo.persistence.Findings) == 0 {
		t.Fatal("typed research result was not persisted")
	}
	if repo.persistence.Run.ApplicationID != nil {
		t.Fatal("task-only research unexpectedly created or linked an application")
	}
}

func TestResearchTaskRejectsCrossUserParentAndCycle(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	repo := newFakeRepository()
	service := newTestService(repo, userID, &fakeApplications{}, &fakeProfiles{})
	foreign := &ResearchTask{Base: Base{ID: uuid.New()}, UserID: uuid.New(), Title: "Foreign", Status: "draft"}
	repo.tasks[foreign.ID] = foreign
	_, err := service.CreateResearchTask(ctx, CreateResearchTaskRequest{UserID: userID, ParentTaskID: &foreign.ID, Title: "Child", TaskType: "general_research"}, false)
	if !errors.Is(err, ErrInvalidResearchTaskParent) {
		t.Fatalf("error = %v", err)
	}
	a := &ResearchTask{Base: Base{ID: uuid.New()}, UserID: userID, Title: "A", Status: "draft"}
	b := &ResearchTask{Base: Base{ID: uuid.New()}, UserID: userID, Title: "B", Status: "draft", ParentTaskID: &a.ID}
	a.ParentTaskID = &b.ID
	repo.tasks[a.ID], repo.tasks[b.ID] = a, b
	if err = service.validateTaskRelationships(ctx, a); !errors.Is(err, ErrResearchTaskParentCycle) {
		t.Fatalf("cycle error = %v", err)
	}
}

func TestResearchTaskQueueCancelAndRetryTransitions(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	repo := newFakeRepository()
	service := newTestService(repo, userID, &fakeApplications{}, &fakeProfiles{})
	task := &ResearchTask{Base: Base{ID: uuid.New()}, UserID: userID, Status: "draft"}
	repo.tasks[task.ID] = task

	if _, err := service.TransitionResearchTask(ctx, task.ID, "queue"); err != nil || task.Status != "queued" {
		t.Fatalf("queue failed: status=%s err=%v", task.Status, err)
	}
	if _, err := service.TransitionResearchTask(ctx, task.ID, "cancel"); err != nil || task.Status != "cancelled" {
		t.Fatalf("cancel failed: status=%s err=%v", task.Status, err)
	}
	if _, err := service.TransitionResearchTask(ctx, task.ID, "retry"); !errors.Is(err, ErrInvalidResearchTaskState) {
		t.Fatalf("cancelled retry error = %v", err)
	}
	failure := "provider unavailable"
	task.Status, task.FailedAt, task.FailureReason = "failed", timePointer(time.Now().UTC()), &failure
	if _, err := service.TransitionResearchTask(ctx, task.ID, "retry"); err != nil || task.Status != "queued" || task.FailedAt != nil || task.FailureReason != nil {
		t.Fatalf("failed retry did not reset failure: status=%s err=%v", task.Status, err)
	}
}

func TestProposalRequiresHumanApprovalAndCannotApproveTwice(t *testing.T) {
	ctx, userID, parentID := context.Background(), uuid.New(), uuid.New()
	repo := newFakeRepository()
	task := &ResearchTask{Base: Base{ID: uuid.New()}, UserID: userID, Status: "review_required"}
	repo.tasks[task.ID] = task
	repo.proposal = &ApplicationProposal{Base: Base{ID: uuid.New()}, UserID: userID, ResearchTaskID: task.ID, Name: "Example", Status: "pending"}
	parent := &profile.ApplicantProfile{Base: profile.Base{ID: parentID}, UserID: userID, ProfileType: profile.ProfileTypeMaster}
	service := newTestService(repo, userID, &fakeApplications{}, &fakeProfiles{item: parent})
	created, err := service.ApproveProposal(ctx, repo.proposal.ID, ApproveProposalRequest{ParentProfileID: parentID})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == uuid.Nil || !repo.approved {
		t.Fatal("approval did not create application")
	}
	if _, err = service.ApproveProposal(ctx, repo.proposal.ID, ApproveProposalRequest{ParentProfileID: parentID}); !errors.Is(err, ErrInvalidProposalState) {
		t.Fatalf("second approval error = %v", err)
	}
}

func TestProposalRejectionDoesNotCreateApplicationAndCanBeReopened(t *testing.T) {
	ctx, userID := context.Background(), uuid.New()
	repo := newFakeRepository()
	task := &ResearchTask{Base: Base{ID: uuid.New()}, UserID: userID, Status: "review_required"}
	repo.tasks[task.ID] = task
	repo.proposal = &ApplicationProposal{Base: Base{ID: uuid.New()}, UserID: userID, ResearchTaskID: task.ID, Name: "Example", Status: "pending"}
	service := newTestService(repo, userID, &fakeApplications{}, &fakeProfiles{})

	proposal, err := service.ReviewProposal(ctx, repo.proposal.ID, "reject")
	if err != nil {
		t.Fatal(err)
	}
	if proposal.Status != "rejected" || repo.approved || task.Status != "completed" {
		t.Fatalf("proposal=%s approved=%v task=%s", proposal.Status, repo.approved, task.Status)
	}
	proposal, err = service.ReviewProposal(ctx, repo.proposal.ID, "reopen")
	if err != nil {
		t.Fatal(err)
	}
	if proposal.Status != "pending" || task.Status != "review_required" {
		t.Fatalf("proposal=%s task=%s", proposal.Status, task.Status)
	}
}

func TestInformationRequestResponseHistoryAndCompletion(t *testing.T) {
	ctx, userID, questionID := context.Background(), uuid.New(), uuid.New()
	repo := newFakeRepository()
	service := newTestService(repo, userID, &fakeApplications{}, &fakeProfiles{})
	item, err := service.CreateInformationRequest(ctx, CreateInformationRequest{UserID: userID, QuestionID: &questionID, RequestType: "missing_information", Title: "Passport expiry", Prompt: "Expiry date?", ResponseType: "date", CreatedBy: "agent"})
	if err != nil {
		t.Fatal(err)
	}
	answer := "2031-12-31"
	if _, err = service.RespondInformationRequest(ctx, item.ID, RespondInformationRequest{ResponseText: &answer, SubmittedBy: "user"}); err != nil {
		t.Fatal(err)
	}
	if item.Status != "answered" || len(repo.history) != 1 {
		t.Fatal("response was not retained")
	}
	if _, err = service.ProcessInformationRequest(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	if item.Status != "completed" {
		t.Fatalf("status = %s", item.Status)
	}
	if _, err = service.ProcessInformationRequest(ctx, item.ID); !errors.Is(err, ErrInvalidInformationState) {
		t.Fatalf("repeat process error = %v", err)
	}
	if _, err = service.TransitionInformationRequest(ctx, item.ID, "reopen"); err != nil || item.Status != "reopened" {
		t.Fatalf("reopen failed: %v", err)
	}
}

func TestPrefillUsesEvidenceAndSkipsCompletedInformationRequest(t *testing.T) {
	ctx, userID, appID, profileID, questionnaireID := context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	institution := "Example University"
	institutionKey, passportKey, motivationKey := "undergraduate_institution", "passport_expiry", "programme_motivation"
	institutionQuestion, passportQuestion, motivationQuestion := uuid.New(), uuid.New(), uuid.New()
	app := &application.Application{Base: application.Base{ID: appID}, UserID: userID, ApplicantProfileID: profileID, Name: "Example"}
	repo := newFakeRepository()
	repo.prefillData = &PrefillData{Application: app, Questionnaires: []QuestionnaireBundle{{Questionnaire: ApplicationQuestionnaire{Base: Base{ID: questionnaireID}, ApplicationID: appID}, Questions: []ApplicationQuestion{{Base: Base{ID: institutionQuestion}, QuestionnaireID: questionnaireID, Key: &institutionKey, Prompt: "Undergraduate institution", Status: "unanswered"}, {Base: Base{ID: passportQuestion}, QuestionnaireID: questionnaireID, Key: &passportKey, Prompt: "Passport expiry", Status: "unanswered"}, {Base: Base{ID: motivationQuestion}, QuestionnaireID: questionnaireID, Key: &motivationKey, Prompt: "Why?", Status: "unanswered"}}, Answers: map[uuid.UUID][]ApplicationAnswer{}}}, CompletedInfo: []InformationRequest{{QuestionID: &passportQuestion, Status: "completed"}}}
	effective := &profile.EffectiveProfileResponse{FullProfileResponse: profile.FullProfileResponse{ProfileResponse: profile.ProfileResponse{ID: profileID}, Education: []profile.SectionResponse{{ID: uuid.New(), Institution: &institution}}}}
	service := newTestService(repo, userID, &fakeApplications{item: app}, &fakeProfiles{effective: effective})
	if _, err := service.RunPrefill(ctx, appID, PrefillRequest{}); err != nil {
		t.Fatal(err)
	}
	if repo.prefill == nil {
		t.Fatal("prefill result not persisted")
	}
	for _, request := range repo.prefill.InformationRequests {
		if request.QuestionID != nil && *request.QuestionID == passportQuestion {
			t.Fatal("completed information request was recreated")
		}
	}
	if len(repo.prefill.Answers) != 2 {
		t.Fatalf("answers = %d, want factual and subjective answers", len(repo.prefill.Answers))
	}
	statuses := map[uuid.UUID]string{}
	for _, answer := range repo.prefill.Answers {
		statuses[answer.QuestionID] = answer.Status
	}
	if statuses[institutionQuestion] != "answered" || statuses[motivationQuestion] != "suggested" {
		t.Fatalf("answer statuses = %#v", statuses)
	}
}
