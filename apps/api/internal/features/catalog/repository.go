package catalog

import (
	"context"

	"github.com/google/uuid"
)

type Filters struct {
	InstitutionID        *uuid.UUID
	Country, DegreeLevel *string
}

type Repository interface {
	ListInstitutions(context.Context) ([]Institution, error)
	CreateInstitution(context.Context, *Institution) error
	GetInstitution(context.Context, uuid.UUID) (*Institution, error)
	UpdateInstitution(context.Context, *Institution) error
	ListProgrammes(context.Context, Filters) ([]Programme, error)
	CreateProgramme(context.Context, *Programme) error
	GetProgramme(context.Context, uuid.UUID) (*Programme, error)
	UpdateProgramme(context.Context, *Programme) error
	ListScholarships(context.Context, Filters) ([]Scholarship, error)
	CreateScholarship(context.Context, *Scholarship) error
	GetScholarship(context.Context, uuid.UUID) (*Scholarship, error)
	UpdateScholarship(context.Context, *Scholarship) error
}
