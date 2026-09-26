package catalog

import (
	"context"
	"github.com/google/uuid"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) ListInstitutions(ctx context.Context) ([]Institution, error) {
	return s.repo.ListInstitutions(ctx)
}
func (s *Service) CreateInstitution(ctx context.Context, r InstitutionRequest) (*Institution, error) {
	v := &Institution{Name: r.Name, ShortName: r.ShortName, InstitutionType: r.InstitutionType, Country: r.Country, City: r.City, WebsiteURL: r.WebsiteURL}
	return v, s.repo.CreateInstitution(ctx, v)
}
func (s *Service) GetInstitution(ctx context.Context, id uuid.UUID) (*Institution, error) {
	return s.repo.GetInstitution(ctx, id)
}
func (s *Service) UpdateInstitution(ctx context.Context, id uuid.UUID, r UpdateInstitutionRequest) (*Institution, error) {
	v, e := s.repo.GetInstitution(ctx, id)
	if e != nil {
		return nil, e
	}
	if r.Name != nil {
		v.Name = *r.Name
	}
	if r.ShortName != nil {
		v.ShortName = r.ShortName
	}
	if r.InstitutionType != nil {
		v.InstitutionType = r.InstitutionType
	}
	if r.Country != nil {
		v.Country = *r.Country
	}
	if r.City != nil {
		v.City = r.City
	}
	if r.WebsiteURL != nil {
		v.WebsiteURL = r.WebsiteURL
	}
	return v, s.repo.UpdateInstitution(ctx, v)
}
func (s *Service) ListProgrammes(ctx context.Context, f Filters) ([]Programme, error) {
	return s.repo.ListProgrammes(ctx, f)
}
func (s *Service) CreateProgramme(ctx context.Context, r ProgrammeRequest) (*Programme, error) {
	if _, e := s.repo.GetInstitution(ctx, r.InstitutionID); e != nil {
		return nil, e
	}
	v := &Programme{InstitutionID: r.InstitutionID, Name: r.Name, DegreeLevel: r.DegreeLevel, FieldOfStudy: r.FieldOfStudy, Faculty: r.Faculty, Department: r.Department, DurationMonths: r.DurationMonths, Mode: r.Mode, Language: r.Language, ProgrammeURL: r.ProgrammeURL, Description: r.Description}
	return v, s.repo.CreateProgramme(ctx, v)
}
func (s *Service) GetProgramme(ctx context.Context, id uuid.UUID) (*Programme, error) {
	return s.repo.GetProgramme(ctx, id)
}
func (s *Service) UpdateProgramme(ctx context.Context, id uuid.UUID, r UpdateProgrammeRequest) (*Programme, error) {
	v, e := s.repo.GetProgramme(ctx, id)
	if e != nil {
		return nil, e
	}
	if r.InstitutionID != nil {
		if _, e = s.repo.GetInstitution(ctx, *r.InstitutionID); e != nil {
			return nil, e
		}
		v.InstitutionID = *r.InstitutionID
	}
	if r.Name != nil {
		v.Name = *r.Name
	}
	if r.DegreeLevel != nil {
		v.DegreeLevel = r.DegreeLevel
	}
	if r.FieldOfStudy != nil {
		v.FieldOfStudy = r.FieldOfStudy
	}
	if r.Faculty != nil {
		v.Faculty = r.Faculty
	}
	if r.Department != nil {
		v.Department = r.Department
	}
	if r.DurationMonths != nil {
		v.DurationMonths = r.DurationMonths
	}
	if r.Mode != nil {
		v.Mode = r.Mode
	}
	if r.Language != nil {
		v.Language = r.Language
	}
	if r.ProgrammeURL != nil {
		v.ProgrammeURL = r.ProgrammeURL
	}
	if r.Description != nil {
		v.Description = r.Description
	}
	return v, s.repo.UpdateProgramme(ctx, v)
}
func (s *Service) ListScholarships(ctx context.Context, f Filters) ([]Scholarship, error) {
	return s.repo.ListScholarships(ctx, f)
}
func (s *Service) CreateScholarship(ctx context.Context, r ScholarshipRequest) (*Scholarship, error) {
	if r.InstitutionID != nil {
		if _, e := s.repo.GetInstitution(ctx, *r.InstitutionID); e != nil {
			return nil, e
		}
	}
	v := &Scholarship{InstitutionID: r.InstitutionID, Name: r.Name, ProviderName: r.ProviderName, Description: r.Description, Country: r.Country, DegreeLevel: r.DegreeLevel, ScholarshipType: r.ScholarshipType, OfficialURL: r.OfficialURL, IsRecurring: r.IsRecurring}
	return v, s.repo.CreateScholarship(ctx, v)
}
func (s *Service) GetScholarship(ctx context.Context, id uuid.UUID) (*Scholarship, error) {
	return s.repo.GetScholarship(ctx, id)
}
func (s *Service) UpdateScholarship(ctx context.Context, id uuid.UUID, r UpdateScholarshipRequest) (*Scholarship, error) {
	v, e := s.repo.GetScholarship(ctx, id)
	if e != nil {
		return nil, e
	}
	if r.InstitutionID != nil {
		if _, e = s.repo.GetInstitution(ctx, *r.InstitutionID); e != nil {
			return nil, e
		}
		v.InstitutionID = r.InstitutionID
	}
	if r.Name != nil {
		v.Name = *r.Name
	}
	if r.ProviderName != nil {
		v.ProviderName = r.ProviderName
	}
	if r.Description != nil {
		v.Description = r.Description
	}
	if r.Country != nil {
		v.Country = r.Country
	}
	if r.DegreeLevel != nil {
		v.DegreeLevel = r.DegreeLevel
	}
	if r.ScholarshipType != nil {
		v.ScholarshipType = r.ScholarshipType
	}
	if r.OfficialURL != nil {
		v.OfficialURL = r.OfficialURL
	}
	if r.IsRecurring != nil {
		v.IsRecurring = *r.IsRecurring
	}
	return v, s.repo.UpdateScholarship(ctx, v)
}
