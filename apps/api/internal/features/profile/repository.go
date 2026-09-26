package profile

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *ApplicantProfile) error
	GetByID(context.Context, uuid.UUID) (*ApplicantProfile, error)
	GetFull(context.Context, uuid.UUID) (*ApplicantProfile, error)
	ListByUserID(context.Context, uuid.UUID) ([]ApplicantProfile, error)
	Update(context.Context, *ApplicantProfile) error
	GetPersonalInfo(context.Context, uuid.UUID) (*ProfilePersonalInfo, error)
	UpsertPersonalInfo(context.Context, *ProfilePersonalInfo) error
}

type BundleRepository interface {
	CreateBundle(context.Context, *ApplicantProfile, *ProfilePersonalInfo, []any, *AuditLog) error
}

type SectionRepository interface {
	ListEducation(context.Context, uuid.UUID) ([]EducationHistory, error)
	GetEducation(context.Context, uuid.UUID, uuid.UUID) (*EducationHistory, error)
	CreateEducation(context.Context, *EducationHistory) error
	UpdateEducation(context.Context, *EducationHistory) error
	DeleteEducation(context.Context, uuid.UUID, uuid.UUID) error
	ListEmployment(context.Context, uuid.UUID) ([]EmploymentHistory, error)
	GetEmployment(context.Context, uuid.UUID, uuid.UUID) (*EmploymentHistory, error)
	CreateEmployment(context.Context, *EmploymentHistory) error
	UpdateEmployment(context.Context, *EmploymentHistory) error
	DeleteEmployment(context.Context, uuid.UUID, uuid.UUID) error
	ListProjects(context.Context, uuid.UUID) ([]Project, error)
	GetProject(context.Context, uuid.UUID, uuid.UUID) (*Project, error)
	CreateProject(context.Context, *Project) error
	UpdateProject(context.Context, *Project) error
	DeleteProject(context.Context, uuid.UUID, uuid.UUID) error
	ListPublications(context.Context, uuid.UUID) ([]Publication, error)
	GetPublication(context.Context, uuid.UUID, uuid.UUID) (*Publication, error)
	CreatePublication(context.Context, *Publication) error
	UpdatePublication(context.Context, *Publication) error
	DeletePublication(context.Context, uuid.UUID, uuid.UUID) error
	ListArticles(context.Context, uuid.UUID) ([]Article, error)
	GetArticle(context.Context, uuid.UUID, uuid.UUID) (*Article, error)
	CreateArticle(context.Context, *Article) error
	UpdateArticle(context.Context, *Article) error
	DeleteArticle(context.Context, uuid.UUID, uuid.UUID) error
	ListSkills(context.Context, uuid.UUID) ([]Skill, error)
	GetSkill(context.Context, uuid.UUID, uuid.UUID) (*Skill, error)
	CreateSkill(context.Context, *Skill) error
	UpdateSkill(context.Context, *Skill) error
	DeleteSkill(context.Context, uuid.UUID, uuid.UUID) error
	ListResearchInterests(context.Context, uuid.UUID) ([]ResearchInterest, error)
	GetResearchInterest(context.Context, uuid.UUID, uuid.UUID) (*ResearchInterest, error)
	CreateResearchInterest(context.Context, *ResearchInterest) error
	UpdateResearchInterest(context.Context, *ResearchInterest) error
	DeleteResearchInterest(context.Context, uuid.UUID, uuid.UUID) error
	ListCareerGoals(context.Context, uuid.UUID) ([]CareerGoal, error)
	GetCareerGoal(context.Context, uuid.UUID, uuid.UUID) (*CareerGoal, error)
	CreateCareerGoal(context.Context, *CareerGoal) error
	UpdateCareerGoal(context.Context, *CareerGoal) error
	DeleteCareerGoal(context.Context, uuid.UUID, uuid.UUID) error
	ListCertifications(context.Context, uuid.UUID) ([]Certification, error)
	GetCertification(context.Context, uuid.UUID, uuid.UUID) (*Certification, error)
	CreateCertification(context.Context, *Certification) error
	UpdateCertification(context.Context, *Certification) error
	DeleteCertification(context.Context, uuid.UUID, uuid.UUID) error
	ListAwards(context.Context, uuid.UUID) ([]Award, error)
	GetAward(context.Context, uuid.UUID, uuid.UUID) (*Award, error)
	CreateAward(context.Context, *Award) error
	UpdateAward(context.Context, *Award) error
	DeleteAward(context.Context, uuid.UUID, uuid.UUID) error
	ListVolunteering(context.Context, uuid.UUID) ([]VolunteerExperience, error)
	GetVolunteer(context.Context, uuid.UUID, uuid.UUID) (*VolunteerExperience, error)
	CreateVolunteer(context.Context, *VolunteerExperience) error
	UpdateVolunteer(context.Context, *VolunteerExperience) error
	DeleteVolunteer(context.Context, uuid.UUID, uuid.UUID) error
}
