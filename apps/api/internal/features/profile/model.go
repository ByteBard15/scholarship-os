package profile

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ApplicantProfile struct {
	Base
	UserID          uuid.UUID   `gorm:"type:uuid;not null;index"`
	Name            string      `gorm:"not null"`
	ProfileType     ProfileType `gorm:"type:text;not null;default:master;index"`
	ParentProfileID *uuid.UUID  `gorm:"type:uuid;index"`
	Headline        *string
	Summary         *string `gorm:"type:text"`
	IsDefault       bool    `gorm:"not null;default:false;index"`
	SoftDelete
	PersonalInfo      *ProfilePersonalInfo  `gorm:"foreignKey:ProfileID"`
	Education         []EducationHistory    `gorm:"foreignKey:ProfileID"`
	Employment        []EmploymentHistory   `gorm:"foreignKey:ProfileID"`
	Projects          []Project             `gorm:"foreignKey:ProfileID"`
	Publications      []Publication         `gorm:"foreignKey:ProfileID"`
	Articles          []Article             `gorm:"foreignKey:ProfileID"`
	Skills            []Skill               `gorm:"foreignKey:ProfileID"`
	ResearchInterests []ResearchInterest    `gorm:"foreignKey:ProfileID"`
	CareerGoals       []CareerGoal          `gorm:"foreignKey:ProfileID"`
	Certifications    []Certification       `gorm:"foreignKey:ProfileID"`
	Awards            []Award               `gorm:"foreignKey:ProfileID"`
	Volunteering      []VolunteerExperience `gorm:"foreignKey:ProfileID"`
}

func (ApplicantProfile) TableName() string { return "applicant_profiles" }

type ProfilePersonalInfo struct {
	Base
	ProfileID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	FirstName     string    `gorm:"not null"`
	MiddleName    *string
	LastName      string `gorm:"not null"`
	PreferredName *string
	Phone         *string
	City          *string
	StateOrRegion *string
	Country       *string
	Nationality   *string
	LinkedInURL   *string
	GitHubURL     *string `gorm:"column:github_url"`
	WebsiteURL    *string
}

func (ProfilePersonalInfo) TableName() string { return "profile_personal_info" }

type EducationHistory struct {
	Base
	ProfileID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Institution    string     `gorm:"not null"`
	Degree         string     `gorm:"not null"`
	FieldOfStudy   string     `gorm:"not null"`
	StartDate      *time.Time `gorm:"type:date"`
	EndDate        *time.Time `gorm:"type:date"`
	IsCurrent      bool
	Grade          *string
	GradeScale     *string
	Classification *string
	City           *string
	Country        *string
	Description    *string `gorm:"type:text"`
	SoftDelete
}

func (EducationHistory) TableName() string { return "education_history" }

type EmploymentHistory struct {
	Base
	ProfileID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Organization   string    `gorm:"not null"`
	JobTitle       string    `gorm:"not null"`
	EmploymentType *string
	StartDate      *time.Time `gorm:"type:date"`
	EndDate        *time.Time `gorm:"type:date"`
	IsCurrent      bool
	City           *string
	Country        *string
	Description    *string `gorm:"type:text"`
	SoftDelete
}

func (EmploymentHistory) TableName() string { return "employment_history" }

type Project struct {
	Base
	ProfileID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Title         string    `gorm:"not null"`
	Organization  *string
	StartDate     *time.Time `gorm:"type:date"`
	EndDate       *time.Time `gorm:"type:date"`
	Description   *string    `gorm:"type:text"`
	URL           *string
	RepositoryURL *string
	SoftDelete
}

func (Project) TableName() string { return "projects" }

type Publication struct {
	Base
	ProfileID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Title           string    `gorm:"not null"`
	PublicationType *string
	Publisher       *string
	PublicationDate *time.Time `gorm:"type:date"`
	DOI             *string
	URL             *string
	Description     *string `gorm:"type:text"`
	SoftDelete
}

func (Publication) TableName() string { return "publications" }

type Article struct {
	Base
	ProfileID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title           string     `gorm:"not null"`
	PublicationDate *time.Time `gorm:"type:date"`
	URL             *string
	Description     *string `gorm:"type:text"`
	SoftDelete
}

func (Article) TableName() string { return "articles" }

type Skill struct {
	Base
	ProfileID       uuid.UUID `gorm:"type:uuid;not null;index:idx_skills_profile_name,unique"`
	Name            string    `gorm:"not null;index:idx_skills_profile_name,unique"`
	Category        *string
	Proficiency     *string
	YearsExperience *float64
	SoftDelete
}

func (Skill) TableName() string { return "skills" }

type ResearchInterest struct {
	Base
	ProfileID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"not null"`
	Description *string   `gorm:"type:text"`
	Priority    *int
	SoftDelete
}

func (ResearchInterest) TableName() string { return "research_interests" }

type CareerGoal struct {
	Base
	ProfileID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text;not null"`
	GoalType    *string
	Priority    *int
	SoftDelete
}

func (CareerGoal) TableName() string { return "career_goals" }

type Certification struct {
	Base
	ProfileID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Name           string    `gorm:"not null"`
	Issuer         *string
	IssueDate      *time.Time `gorm:"type:date"`
	ExpirationDate *time.Time `gorm:"type:date"`
	CredentialID   *string
	URL            *string
	SoftDelete
}

func (Certification) TableName() string { return "certifications" }

type Award struct {
	Base
	ProfileID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Title       string    `gorm:"not null"`
	Issuer      *string
	AwardDate   *time.Time `gorm:"type:date"`
	Description *string    `gorm:"type:text"`
	SoftDelete
}

func (Award) TableName() string { return "awards" }

type VolunteerExperience struct {
	Base
	ProfileID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Organization string     `gorm:"not null"`
	Role         string     `gorm:"not null"`
	StartDate    *time.Time `gorm:"type:date"`
	EndDate      *time.Time `gorm:"type:date"`
	IsCurrent    bool
	Description  *string `gorm:"type:text"`
	SoftDelete
}

func (VolunteerExperience) TableName() string { return "volunteer_experiences" }
