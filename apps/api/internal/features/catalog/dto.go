package catalog

import (
	"time"

	"github.com/google/uuid"
)

type InstitutionRequest struct {
	Name            string  `json:"name" validate:"required,max=250"`
	ShortName       *string `json:"shortName"`
	InstitutionType *string `json:"institutionType"`
	Country         string  `json:"country" validate:"required,max=120"`
	City            *string `json:"city"`
	WebsiteURL      *string `json:"websiteUrl" validate:"omitempty,url"`
}
type UpdateInstitutionRequest struct {
	Name            *string `json:"name" validate:"omitempty,min=1,max=250"`
	ShortName       *string `json:"shortName"`
	InstitutionType *string `json:"institutionType"`
	Country         *string `json:"country" validate:"omitempty,min=1,max=120"`
	City            *string `json:"city"`
	WebsiteURL      *string `json:"websiteUrl" validate:"omitempty,url"`
}

type ProgrammeRequest struct {
	InstitutionID  uuid.UUID `json:"institutionId" validate:"required"`
	Name           string    `json:"name" validate:"required,max=250"`
	DegreeLevel    *string   `json:"degreeLevel"`
	FieldOfStudy   *string   `json:"fieldOfStudy"`
	Faculty        *string   `json:"faculty"`
	Department     *string   `json:"department"`
	DurationMonths *int      `json:"durationMonths" validate:"omitempty,gt=0"`
	Mode           *string   `json:"mode"`
	Language       *string   `json:"language"`
	ProgrammeURL   *string   `json:"programmeUrl" validate:"omitempty,url"`
	Description    *string   `json:"description"`
}
type UpdateProgrammeRequest struct {
	InstitutionID  *uuid.UUID `json:"institutionId"`
	Name           *string    `json:"name" validate:"omitempty,min=1,max=250"`
	DegreeLevel    *string    `json:"degreeLevel"`
	FieldOfStudy   *string    `json:"fieldOfStudy"`
	Faculty        *string    `json:"faculty"`
	Department     *string    `json:"department"`
	DurationMonths *int       `json:"durationMonths" validate:"omitempty,gt=0"`
	Mode           *string    `json:"mode"`
	Language       *string    `json:"language"`
	ProgrammeURL   *string    `json:"programmeUrl" validate:"omitempty,url"`
	Description    *string    `json:"description"`
}

type ScholarshipRequest struct {
	InstitutionID   *uuid.UUID `json:"institutionId"`
	Name            string     `json:"name" validate:"required,max=250"`
	ProviderName    *string    `json:"providerName"`
	Description     *string    `json:"description"`
	Country         *string    `json:"country"`
	DegreeLevel     *string    `json:"degreeLevel"`
	ScholarshipType *string    `json:"scholarshipType"`
	OfficialURL     *string    `json:"officialUrl" validate:"omitempty,url"`
	IsRecurring     bool       `json:"isRecurring"`
}
type UpdateScholarshipRequest struct {
	InstitutionID   *uuid.UUID `json:"institutionId"`
	Name            *string    `json:"name" validate:"omitempty,min=1,max=250"`
	ProviderName    *string    `json:"providerName"`
	Description     *string    `json:"description"`
	Country         *string    `json:"country"`
	DegreeLevel     *string    `json:"degreeLevel"`
	ScholarshipType *string    `json:"scholarshipType"`
	OfficialURL     *string    `json:"officialUrl" validate:"omitempty,url"`
	IsRecurring     *bool      `json:"isRecurring"`
}

type InstitutionResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	ShortName       *string   `json:"shortName,omitempty"`
	InstitutionType *string   `json:"institutionType,omitempty"`
	Country         string    `json:"country"`
	City            *string   `json:"city,omitempty"`
	WebsiteURL      *string   `json:"websiteUrl,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
type ProgrammeResponse struct {
	ID             uuid.UUID `json:"id"`
	InstitutionID  uuid.UUID `json:"institutionId"`
	Name           string    `json:"name"`
	DegreeLevel    *string   `json:"degreeLevel,omitempty"`
	FieldOfStudy   *string   `json:"fieldOfStudy,omitempty"`
	Faculty        *string   `json:"faculty,omitempty"`
	Department     *string   `json:"department,omitempty"`
	DurationMonths *int      `json:"durationMonths,omitempty"`
	Mode           *string   `json:"mode,omitempty"`
	Language       *string   `json:"language,omitempty"`
	ProgrammeURL   *string   `json:"programmeUrl,omitempty"`
	Description    *string   `json:"description,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
type ScholarshipResponse struct {
	ID              uuid.UUID  `json:"id"`
	InstitutionID   *uuid.UUID `json:"institutionId,omitempty"`
	Name            string     `json:"name"`
	ProviderName    *string    `json:"providerName,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Country         *string    `json:"country,omitempty"`
	DegreeLevel     *string    `json:"degreeLevel,omitempty"`
	ScholarshipType *string    `json:"scholarshipType,omitempty"`
	OfficialURL     *string    `json:"officialUrl,omitempty"`
	IsRecurring     bool       `json:"isRecurring"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func toInstitution(v *Institution) InstitutionResponse {
	return InstitutionResponse{v.ID, v.Name, v.ShortName, v.InstitutionType, v.Country, v.City, v.WebsiteURL, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}
func toProgramme(v *Programme) ProgrammeResponse {
	return ProgrammeResponse{v.ID, v.InstitutionID, v.Name, v.DegreeLevel, v.FieldOfStudy, v.Faculty, v.Department, v.DurationMonths, v.Mode, v.Language, v.ProgrammeURL, v.Description, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}
func toScholarship(v *Scholarship) ScholarshipResponse {
	return ScholarshipResponse{v.ID, v.InstitutionID, v.Name, v.ProviderName, v.Description, v.Country, v.DegreeLevel, v.ScholarshipType, v.OfficialURL, v.IsRecurring, v.CreatedAt.UTC(), v.UpdatedAt.UTC()}
}
