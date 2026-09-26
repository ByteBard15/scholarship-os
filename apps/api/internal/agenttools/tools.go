// Package agenttools exposes the narrow, typed operations available to future
// scholarship agents. It intentionally contains no database handle or generic
// persistence method.
package agenttools

import (
	"context"

	"github.com/example/scholarship-os/apps/api/internal/features/application"
	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
)

type ScholarshipAgentTools interface {
	GetApplication(context.Context, uuid.UUID) (*ApplicationView, error)
	GetEffectiveProfile(context.Context, uuid.UUID) (*profile.EffectiveProfileResponse, error)
	CreateResearchRun(context.Context, uuid.UUID, application.ResearchRunRequest) (*ResearchRunView, error)
	AddResearchSource(context.Context, uuid.UUID, uuid.UUID, application.ResearchSourceCandidate) (*ResearchSourceView, error)
	AddResearchFinding(context.Context, uuid.UUID, uuid.UUID, application.FindingProposal) (*ResearchFindingView, error)
	ProposeRequirement(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, application.RequirementRequest) (*ResearchFindingView, error)
	ProposeDeadline(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, application.DeadlineRequest) (*ResearchFindingView, error)
	ProposeFunding(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, application.FundingRequest) (*ResearchFindingView, error)
	ProposeContact(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, application.ContactRequest) (*ResearchFindingView, error)
	ProposeSupervisor(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, application.SupervisorRequest) (*ResearchFindingView, error)
	CreateApplicationTask(context.Context, uuid.UUID, application.TaskRequest) (*TaskView, error)
	UpdateApplicationTask(context.Context, uuid.UUID, uuid.UUID, application.TaskRequest) (*TaskView, error)
}

type ApplicationView struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ApplicantProfileID uuid.UUID
	Name               string
	Status             application.Status
	ResearchStatus     application.ResearchStatus
	InstitutionName    *string
	ProgrammeName      *string
	ScholarshipName    *string
}
type ResearchRunView struct {
	ID            uuid.UUID
	ApplicationID *uuid.UUID
	Status        string
	ResearchType  string
}
type ResearchSourceView struct {
	ID            uuid.UUID
	ResearchRunID uuid.UUID
	URL           string
	SourceType    string
	IsOfficial    bool
}
type ResearchFindingView struct {
	ID                 uuid.UUID
	ResearchRunID      uuid.UUID
	SourceID           *uuid.UUID
	Category           string
	Field              string
	ReviewStatus       string
	VerificationStatus string
}
type TaskView struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
	ParentTaskID  *uuid.UUID
	Title         string
	Status        string
}

type applicationService interface {
	Get(context.Context, uuid.UUID) (*application.Application, error)
	CreateTask(context.Context, uuid.UUID, application.TaskRequest) (*application.ApplicationTask, error)
	UpdateTask(context.Context, uuid.UUID, uuid.UUID, application.TaskRequest) (*application.ApplicationTask, error)
}
type profileService interface {
	ResolveEffectiveProfile(context.Context, uuid.UUID) (*profile.EffectiveProfileResponse, error)
}
type researchService interface {
	Run(context.Context, uuid.UUID, application.ResearchRunRequest) (*application.ResearchRun, error)
	AddSource(context.Context, uuid.UUID, uuid.UUID, application.ResearchSourceCandidate) (*application.ResearchSource, error)
	AddFinding(context.Context, uuid.UUID, uuid.UUID, application.FindingProposal) (*application.ResearchFinding, error)
}

type Tools struct {
	applications applicationService
	profiles     profileService
	research     researchService
}

func New(apps applicationService, profiles profileService, research researchService) *Tools {
	return &Tools{apps, profiles, research}
}
func (t *Tools) GetApplication(c context.Context, id uuid.UUID) (*ApplicationView, error) {
	v, err := t.applications.Get(c, id)
	if err != nil {
		return nil, err
	}
	out := &ApplicationView{ID: v.ID, UserID: v.UserID, ApplicantProfileID: v.ApplicantProfileID, Name: v.Name, Status: v.Status, ResearchStatus: v.ResearchStatus}
	if v.Institution != nil {
		out.InstitutionName = &v.Institution.Name
	}
	if v.Programme != nil {
		out.ProgrammeName = &v.Programme.Name
	}
	if v.Scholarship != nil {
		out.ScholarshipName = &v.Scholarship.Name
	}
	return out, nil
}
func (t *Tools) GetEffectiveProfile(c context.Context, id uuid.UUID) (*profile.EffectiveProfileResponse, error) {
	return t.profiles.ResolveEffectiveProfile(c, id)
}
func (t *Tools) CreateResearchRun(c context.Context, id uuid.UUID, q application.ResearchRunRequest) (*ResearchRunView, error) {
	v, e := t.research.Run(c, id, q)
	if e != nil {
		return nil, e
	}
	return &ResearchRunView{v.ID, v.ApplicationID, v.Status, v.ResearchType}, nil
}
func (t *Tools) AddResearchSource(c context.Context, appID, runID uuid.UUID, q application.ResearchSourceCandidate) (*ResearchSourceView, error) {
	v, e := t.research.AddSource(c, appID, runID, q)
	if e != nil {
		return nil, e
	}
	return &ResearchSourceView{v.ID, v.ResearchRunID, v.URL, v.SourceType, v.IsOfficial}, nil
}
func (t *Tools) AddResearchFinding(c context.Context, appID, runID uuid.UUID, q application.FindingProposal) (*ResearchFindingView, error) {
	v, e := t.research.AddFinding(c, appID, runID, q)
	if e != nil {
		return nil, e
	}
	return findingView(v), nil
}
func (t *Tools) ProposeRequirement(c context.Context, appID, runID uuid.UUID, sourceID *uuid.UUID, q application.RequirementRequest) (*ResearchFindingView, error) {
	return t.AddResearchFinding(c, appID, runID, application.FindingProposal{SourceID: sourceID, Category: "requirement", Field: q.Title, Value: q, VerificationStatus: "unverified"})
}
func (t *Tools) ProposeDeadline(c context.Context, appID, runID uuid.UUID, sourceID *uuid.UUID, q application.DeadlineRequest) (*ResearchFindingView, error) {
	return t.AddResearchFinding(c, appID, runID, application.FindingProposal{SourceID: sourceID, Category: "deadline", Field: q.DeadlineType, Value: q, VerificationStatus: "unverified"})
}
func (t *Tools) ProposeFunding(c context.Context, appID, runID uuid.UUID, sourceID *uuid.UUID, q application.FundingRequest) (*ResearchFindingView, error) {
	return t.AddResearchFinding(c, appID, runID, application.FindingProposal{SourceID: sourceID, Category: "funding", Field: q.FundingType, Value: q, VerificationStatus: "unverified"})
}
func (t *Tools) ProposeContact(c context.Context, appID, runID uuid.UUID, sourceID *uuid.UUID, q application.ContactRequest) (*ResearchFindingView, error) {
	field := "contact"
	if q.ContactType != nil {
		field = *q.ContactType
	}
	return t.AddResearchFinding(c, appID, runID, application.FindingProposal{SourceID: sourceID, Category: "contact", Field: field, Value: q, VerificationStatus: "unverified"})
}
func (t *Tools) ProposeSupervisor(c context.Context, appID, runID uuid.UUID, sourceID *uuid.UUID, q application.SupervisorRequest) (*ResearchFindingView, error) {
	return t.AddResearchFinding(c, appID, runID, application.FindingProposal{SourceID: sourceID, Category: "supervisor", Field: q.Name, Value: q, VerificationStatus: "unverified"})
}
func (t *Tools) CreateApplicationTask(c context.Context, id uuid.UUID, q application.TaskRequest) (*TaskView, error) {
	v, e := t.applications.CreateTask(c, id, q)
	if e != nil {
		return nil, e
	}
	return taskView(v), nil
}
func (t *Tools) UpdateApplicationTask(c context.Context, a, id uuid.UUID, q application.TaskRequest) (*TaskView, error) {
	v, e := t.applications.UpdateTask(c, a, id, q)
	if e != nil {
		return nil, e
	}
	return taskView(v), nil
}
func findingView(v *application.ResearchFinding) *ResearchFindingView {
	return &ResearchFindingView{v.ID, v.ResearchRunID, v.SourceID, v.Category, v.Field, v.ReviewStatus, v.VerificationStatus}
}
func taskView(v *application.ApplicationTask) *TaskView {
	return &TaskView{v.ID, v.ApplicationID, v.ParentTaskID, v.Title, v.Status}
}
