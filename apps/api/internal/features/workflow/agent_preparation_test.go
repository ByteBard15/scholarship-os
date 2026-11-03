package workflow

import (
	"context"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type preparationRepository struct {
	*fakeRepository
	questionnaire *ApplicationQuestionnaire
	question      *ApplicationQuestion
}

func (r *preparationRepository) CreateQuestionnaire(_ context.Context, q *ApplicationQuestionnaire) error {
	q.ID = uuid.New()
	r.questionnaire = q
	return nil
}
func (r *preparationRepository) GetQuestionnaireByID(_ context.Context, id uuid.UUID) (*ApplicationQuestionnaire, error) {
	if r.questionnaire == nil || r.questionnaire.ID != id {
		return nil, ErrQuestionnaireNotFound
	}
	return r.questionnaire, nil
}
func (r *preparationRepository) CreateQuestion(_ context.Context, q *ApplicationQuestion) error {
	r.question = q
	return nil
}
func TestAgentQuestionnaireConstructionCannotClaimCompletion(t *testing.T) {
	owner, appID := uuid.New(), uuid.New()
	r := &preparationRepository{fakeRepository: newFakeRepository()}
	s := NewService(r, nil, &fakeApplications{item: &application.Application{Base: application.Base{ID: appID}, UserID: owner}}, nil, nil)
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: map[string]struct{}{principal.ScopeQuestionnaires: {}}})
	completed := "completed"
	q, err := s.CreateQuestionnaire(ctx, appID, QuestionnaireRequest{Title: "Application questions", Status: &completed})
	if err != nil || q.Status != "not_started" {
		t.Fatal("agent completion accepted", err)
	}
	question, err := s.CreateQuestion(ctx, q.ID, QuestionRequest{Prompt: "Research interests?", Status: &completed})
	if err != nil || question.Status != "unanswered" {
		t.Fatal("agent answer approval accepted", err)
	}
	for _, test := range []struct {
		owner  uuid.UUID
		scopes map[string]struct{}
	}{
		{uuid.New(), map[string]struct{}{principal.ScopeQuestionnaires: {}}}, {owner, map[string]struct{}{}},
	} {
		other := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, UserID: &test.owner, Scopes: test.scopes})
		if _, err := s.CreateQuestionnaire(other, appID, QuestionnaireRequest{Title: "No"}); err == nil {
			t.Fatal("questionnaire guard bypassed")
		}
		if _, err := s.CreateQuestion(other, q.ID, QuestionRequest{Prompt: "No"}); err == nil {
			t.Fatal("question guard bypassed")
		}
	}
}

func TestAgentInformationListingRequiresWorkspace(t *testing.T) {
	service := NewService(nil, nil, nil, nil, nil)
	ctx := principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, Scopes: map[string]struct{}{principal.ScopeInformationRequests: {}}})
	if _, err := service.ListInformationRequests(ctx, InformationRequestFilters{}); err != principal.ErrForbidden {
		t.Fatal("workspace-less information listing", err)
	}
}
