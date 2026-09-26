package application

import (
	"context"
	"time"

	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
)

type Filters struct {
	Status     *Status
	Country    *string
	IntakeYear *int
	UserID     *uuid.UUID
}

type Repository interface {
	CreateWithProfile(context.Context, *profile.ApplicantProfile, *Application, []ApplicationTask, *profile.AuditLog) error
	List(context.Context, Filters) ([]Application, error)
	Get(context.Context, uuid.UUID) (*Application, error)
	GetDetail(context.Context, uuid.UUID) (*ApplicationDetail, error)
	Update(context.Context, *Application, *profile.AuditLog) error
	Delete(context.Context, *Application) error

	ListRequirements(context.Context, uuid.UUID) ([]ApplicationRequirement, error)
	GetRequirement(context.Context, uuid.UUID, uuid.UUID) (*ApplicationRequirement, error)
	CreateRequirement(context.Context, *ApplicationRequirement, *profile.AuditLog) error
	UpdateRequirement(context.Context, *ApplicationRequirement, *profile.AuditLog) error
	DeleteRequirement(context.Context, *ApplicationRequirement) error
	ListDeadlines(context.Context, uuid.UUID) ([]ApplicationDeadline, error)
	GetDeadline(context.Context, uuid.UUID, uuid.UUID) (*ApplicationDeadline, error)
	CreateDeadline(context.Context, *ApplicationDeadline, *profile.AuditLog) error
	UpdateDeadline(context.Context, *ApplicationDeadline, *profile.AuditLog) error
	DeleteDeadline(context.Context, *ApplicationDeadline) error
	UpcomingDeadlines(context.Context, time.Time, time.Time, *uuid.UUID, *uuid.UUID, *Status) ([]ApplicationDeadline, error)
	ListFunding(context.Context, uuid.UUID) ([]ApplicationFunding, error)
	GetFunding(context.Context, uuid.UUID, uuid.UUID) (*ApplicationFunding, error)
	CreateFunding(context.Context, *ApplicationFunding) error
	UpdateFunding(context.Context, *ApplicationFunding) error
	DeleteFunding(context.Context, *ApplicationFunding) error
	ListContacts(context.Context, uuid.UUID) ([]ApplicationContact, error)
	GetContact(context.Context, uuid.UUID, uuid.UUID) (*ApplicationContact, error)
	CreateContact(context.Context, *ApplicationContact) error
	UpdateContact(context.Context, *ApplicationContact) error
	DeleteContact(context.Context, *ApplicationContact) error
	ListSupervisors(context.Context, uuid.UUID) ([]ApplicationSupervisor, error)
	GetSupervisor(context.Context, uuid.UUID, uuid.UUID) (*ApplicationSupervisor, error)
	CreateSupervisor(context.Context, *ApplicationSupervisor) error
	UpdateSupervisor(context.Context, *ApplicationSupervisor) error
	DeleteSupervisor(context.Context, *ApplicationSupervisor) error
	ListURLs(context.Context, uuid.UUID) ([]ApplicationURL, error)
	GetURL(context.Context, uuid.UUID, uuid.UUID) (*ApplicationURL, error)
	CreateURL(context.Context, *ApplicationURL) error
	DeleteURL(context.Context, *ApplicationURL) error
	ListTasks(context.Context, uuid.UUID) ([]ApplicationTask, error)
	GetTask(context.Context, uuid.UUID, uuid.UUID) (*ApplicationTask, error)
	CreateTask(context.Context, *ApplicationTask) error
	UpdateTask(context.Context, *ApplicationTask, *profile.AuditLog) error
	DeleteTask(context.Context, *ApplicationTask) error
	ListEvidence(context.Context, uuid.UUID) ([]RequirementEvidence, error)
	CreateEvidence(context.Context, *RequirementEvidence, *profile.AuditLog) error
	Dashboard(context.Context, *uuid.UUID) (*DashboardData, error)
}

type ApplicationDetail struct {
	Application       *Application
	Requirements      []ApplicationRequirement
	Deadlines         []ApplicationDeadline
	Funding           []ApplicationFunding
	Contacts          []ApplicationContact
	Supervisors       []ApplicationSupervisor
	URLs              []ApplicationURL
	Tasks             []ApplicationTask
	LatestResearchRun *ResearchRun
}
type DashboardData struct {
	Applications           []Application
	Deadlines              []ApplicationDeadline
	Tasks                  []ApplicationTask
	PendingResearch        int64
	ResearchConflicts      int64
	IncompleteRequirements int64
}

type ResearchRepository interface {
	CreateRun(context.Context, *ResearchRun, *Application) error
	PersistResult(context.Context, *ResearchRun, *Application, []ResearchSource, []ResearchFinding) error
	FailRun(context.Context, *ResearchRun, *Application, error) error
	ListRuns(context.Context, uuid.UUID) ([]ResearchRun, error)
	GetRun(context.Context, uuid.UUID, uuid.UUID) (*ResearchRun, error)
	ListSources(context.Context, uuid.UUID) ([]ResearchSource, error)
	CreateSource(context.Context, *ResearchSource) error
	ListFindings(context.Context, uuid.UUID) ([]ResearchFinding, error)
	CreateFinding(context.Context, *ResearchFinding) error
	GetFinding(context.Context, uuid.UUID, uuid.UUID) (*ResearchFinding, error)
	UpdateFinding(context.Context, *ResearchFinding) error
	ApplyFindings(context.Context, *ResearchRun, *Application, AppliedResearch, *profile.AuditLog) error
	MarkStale(context.Context, *Application) error
}

type AppliedResearch struct {
	Requirements []ApplicationRequirement
	Deadlines    []ApplicationDeadline
	Funding      []ApplicationFunding
	Contacts     []ApplicationContact
	Supervisors  []ApplicationSupervisor
	URLs         []ApplicationURL
}
