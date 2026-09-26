package profile

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct{ db *gorm.DB }

func NewGORMRepository(db *gorm.DB) *GORMRepository { return &GORMRepository{db: db} }
func (r *GORMRepository) Create(ctx context.Context, p *ApplicantProfile) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if p.IsDefault {
			if err := tx.Model(&ApplicantProfile{}).Where("user_id = ? AND is_default = ?", p.UserID, true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(p).Error
	})
}
func (r *GORMRepository) GetByID(ctx context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	var p ApplicantProfile
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileNotFound
	}
	return &p, err
}
func (r *GORMRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]ApplicantProfile, error) {
	var v []ApplicantProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&v).Error
	return v, err
}
func (r *GORMRepository) Update(ctx context.Context, p *ApplicantProfile) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if p.IsDefault {
			if err := tx.Model(&ApplicantProfile{}).Where("user_id = ? AND id <> ?", p.UserID, p.ID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(p).Error
	})
}
func (r *GORMRepository) GetFull(ctx context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	var p ApplicantProfile
	err := r.db.WithContext(ctx).Preload("PersonalInfo").Preload("Education").Preload("Employment").Preload("Projects").Preload("Publications").Preload("Articles").Preload("Skills").Preload("ResearchInterests").Preload("CareerGoals").Preload("Certifications").Preload("Awards").Preload("Volunteering").First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileNotFound
	}
	return &p, err
}
func (r *GORMRepository) GetPersonalInfo(ctx context.Context, pid uuid.UUID) (*ProfilePersonalInfo, error) {
	var v ProfilePersonalInfo
	err := r.db.WithContext(ctx).First(&v, "profile_id = ?", pid).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPersonalInfoNotFound
	}
	return &v, err
}
func (r *GORMRepository) UpsertPersonalInfo(ctx context.Context, v *ProfilePersonalInfo) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "profile_id"}}, DoUpdates: clause.AssignmentColumns([]string{"first_name", "middle_name", "last_name", "preferred_name", "phone", "city", "state_or_region", "country", "nationality", "linked_in_url", "github_url", "website_url", "updated_at"})}).Create(v).Error
}

func list(ctx context.Context, db *gorm.DB, pid uuid.UUID, out any) error {
	return db.WithContext(ctx).Where("profile_id = ?", pid).Order("created_at ASC").Find(out).Error
}
func get(ctx context.Context, db *gorm.DB, pid, id uuid.UUID, out any) error {
	err := db.WithContext(ctx).First(out, "profile_id = ? AND id = ?", pid, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSectionNotFound
	}
	return err
}
func create(ctx context.Context, db *gorm.DB, v any) error {
	err := db.WithContext(ctx).Create(v).Error
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return ErrSkillConflict
	}
	return err
}
func update(ctx context.Context, db *gorm.DB, v any) error {
	err := db.WithContext(ctx).Save(v).Error
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return ErrSkillConflict
	}
	return err
}
func remove(ctx context.Context, db *gorm.DB, pid, id uuid.UUID, model any) error {
	result := db.WithContext(ctx).Where("profile_id = ? AND id = ?", pid, id).Delete(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSectionNotFound
	}
	return nil
}

func (r *GORMRepository) ListEducation(c context.Context, p uuid.UUID) ([]EducationHistory, error) {
	var v []EducationHistory
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetEducation(c context.Context, p, id uuid.UUID) (*EducationHistory, error) {
	v := new(EducationHistory)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateEducation(c context.Context, v *EducationHistory) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateEducation(c context.Context, v *EducationHistory) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteEducation(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &EducationHistory{})
}
func (r *GORMRepository) ListEmployment(c context.Context, p uuid.UUID) ([]EmploymentHistory, error) {
	var v []EmploymentHistory
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetEmployment(c context.Context, p, id uuid.UUID) (*EmploymentHistory, error) {
	v := new(EmploymentHistory)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateEmployment(c context.Context, v *EmploymentHistory) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateEmployment(c context.Context, v *EmploymentHistory) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteEmployment(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &EmploymentHistory{})
}
func (r *GORMRepository) ListProjects(c context.Context, p uuid.UUID) ([]Project, error) {
	var v []Project
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetProject(c context.Context, p, id uuid.UUID) (*Project, error) {
	v := new(Project)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateProject(c context.Context, v *Project) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateProject(c context.Context, v *Project) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteProject(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &Project{})
}
func (r *GORMRepository) ListPublications(c context.Context, p uuid.UUID) ([]Publication, error) {
	var v []Publication
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetPublication(c context.Context, p, id uuid.UUID) (*Publication, error) {
	v := new(Publication)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreatePublication(c context.Context, v *Publication) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdatePublication(c context.Context, v *Publication) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeletePublication(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &Publication{})
}
func (r *GORMRepository) ListArticles(c context.Context, p uuid.UUID) ([]Article, error) {
	var v []Article
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetArticle(c context.Context, p, id uuid.UUID) (*Article, error) {
	v := new(Article)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateArticle(c context.Context, v *Article) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateArticle(c context.Context, v *Article) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteArticle(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &Article{})
}
func (r *GORMRepository) ListSkills(c context.Context, p uuid.UUID) ([]Skill, error) {
	var v []Skill
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetSkill(c context.Context, p, id uuid.UUID) (*Skill, error) {
	v := new(Skill)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateSkill(c context.Context, v *Skill) error { return create(c, r.db, v) }
func (r *GORMRepository) UpdateSkill(c context.Context, v *Skill) error { return update(c, r.db, v) }
func (r *GORMRepository) DeleteSkill(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &Skill{})
}
func (r *GORMRepository) ListResearchInterests(c context.Context, p uuid.UUID) ([]ResearchInterest, error) {
	var v []ResearchInterest
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetResearchInterest(c context.Context, p, id uuid.UUID) (*ResearchInterest, error) {
	v := new(ResearchInterest)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateResearchInterest(c context.Context, v *ResearchInterest) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateResearchInterest(c context.Context, v *ResearchInterest) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteResearchInterest(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &ResearchInterest{})
}
func (r *GORMRepository) ListCareerGoals(c context.Context, p uuid.UUID) ([]CareerGoal, error) {
	var v []CareerGoal
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetCareerGoal(c context.Context, p, id uuid.UUID) (*CareerGoal, error) {
	v := new(CareerGoal)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateCareerGoal(c context.Context, v *CareerGoal) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateCareerGoal(c context.Context, v *CareerGoal) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteCareerGoal(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &CareerGoal{})
}
func (r *GORMRepository) ListCertifications(c context.Context, p uuid.UUID) ([]Certification, error) {
	var v []Certification
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetCertification(c context.Context, p, id uuid.UUID) (*Certification, error) {
	v := new(Certification)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateCertification(c context.Context, v *Certification) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateCertification(c context.Context, v *Certification) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteCertification(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &Certification{})
}
func (r *GORMRepository) ListAwards(c context.Context, p uuid.UUID) ([]Award, error) {
	var v []Award
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetAward(c context.Context, p, id uuid.UUID) (*Award, error) {
	v := new(Award)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateAward(c context.Context, v *Award) error { return create(c, r.db, v) }
func (r *GORMRepository) UpdateAward(c context.Context, v *Award) error { return update(c, r.db, v) }
func (r *GORMRepository) DeleteAward(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &Award{})
}
func (r *GORMRepository) ListVolunteering(c context.Context, p uuid.UUID) ([]VolunteerExperience, error) {
	var v []VolunteerExperience
	e := list(c, r.db, p, &v)
	return v, e
}
func (r *GORMRepository) GetVolunteer(c context.Context, p, id uuid.UUID) (*VolunteerExperience, error) {
	v := new(VolunteerExperience)
	e := get(c, r.db, p, id, v)
	return v, e
}
func (r *GORMRepository) CreateVolunteer(c context.Context, v *VolunteerExperience) error {
	return create(c, r.db, v)
}
func (r *GORMRepository) UpdateVolunteer(c context.Context, v *VolunteerExperience) error {
	return update(c, r.db, v)
}
func (r *GORMRepository) DeleteVolunteer(c context.Context, p, id uuid.UUID) error {
	return remove(c, r.db, p, id, &VolunteerExperience{})
}
