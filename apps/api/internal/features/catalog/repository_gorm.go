package catalog

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMRepository struct{ db *gorm.DB }

func NewGORMRepository(db *gorm.DB) *GORMRepository { return &GORMRepository{db: db} }
func (r *GORMRepository) ListInstitutions(ctx context.Context) (v []Institution, err error) {
	err = r.db.WithContext(ctx).Order("name").Find(&v).Error
	return
}
func (r *GORMRepository) CreateInstitution(ctx context.Context, v *Institution) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetInstitution(ctx context.Context, id uuid.UUID) (*Institution, error) {
	var v Institution
	if err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInstitutionNotFound
		}
		return nil, err
	}
	return &v, nil
}
func (r *GORMRepository) UpdateInstitution(ctx context.Context, v *Institution) error {
	return r.db.WithContext(ctx).Save(v).Error
}
func (r *GORMRepository) ListProgrammes(ctx context.Context, f Filters) (v []Programme, err error) {
	q := r.db.WithContext(ctx)
	if f.InstitutionID != nil {
		q = q.Where("institution_id = ?", *f.InstitutionID)
	}
	err = q.Order("name").Find(&v).Error
	return
}
func (r *GORMRepository) CreateProgramme(ctx context.Context, v *Programme) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetProgramme(ctx context.Context, id uuid.UUID) (*Programme, error) {
	var v Programme
	if err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProgrammeNotFound
		}
		return nil, err
	}
	return &v, nil
}
func (r *GORMRepository) UpdateProgramme(ctx context.Context, v *Programme) error {
	return r.db.WithContext(ctx).Save(v).Error
}
func (r *GORMRepository) ListScholarships(ctx context.Context, f Filters) (v []Scholarship, err error) {
	q := r.db.WithContext(ctx)
	if f.InstitutionID != nil {
		q = q.Where("institution_id = ?", *f.InstitutionID)
	}
	if f.Country != nil {
		q = q.Where("country = ?", *f.Country)
	}
	if f.DegreeLevel != nil {
		q = q.Where("degree_level = ?", *f.DegreeLevel)
	}
	err = q.Order("name").Find(&v).Error
	return
}
func (r *GORMRepository) CreateScholarship(ctx context.Context, v *Scholarship) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetScholarship(ctx context.Context, id uuid.UUID) (*Scholarship, error) {
	var v Scholarship
	if err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScholarshipNotFound
		}
		return nil, err
	}
	return &v, nil
}
func (r *GORMRepository) UpdateScholarship(ctx context.Context, v *Scholarship) error {
	return r.db.WithContext(ctx).Save(v).Error
}
