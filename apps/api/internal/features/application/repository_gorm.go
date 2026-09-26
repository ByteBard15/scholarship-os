package application

import (
	"context"
	"errors"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct{ db *gorm.DB }

func NewGORMRepository(db *gorm.DB) *GORMRepository { return &GORMRepository{db} }

func (r *GORMRepository) CreateWithProfile(ctx context.Context, p *profile.ApplicantProfile, a *Application, tasks []ApplicationTask, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(p).Error; e != nil {
			return e
		}
		a.ApplicantProfileID = p.ID
		if e := tx.Create(a).Error; e != nil {
			return e
		}
		for i := range tasks {
			tasks[i].ApplicationID = a.ID
		}
		if len(tasks) > 0 {
			if e := tx.Create(&tasks).Error; e != nil {
				return e
			}
		}
		audit.ProfileID = &p.ID
		audit.EntityID = &a.ID
		return tx.Create(audit).Error
	})
}
func (r *GORMRepository) List(ctx context.Context, f Filters) (v []Application, err error) {
	q := r.db.WithContext(ctx).Preload("Institution").Preload("Programme").Preload("Scholarship")
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.Country != nil {
		q = q.Where("country = ?", *f.Country)
	}
	if f.IntakeYear != nil {
		q = q.Where("intake_year = ?", *f.IntakeYear)
	}
	err = q.Order("updated_at DESC").Find(&v).Error
	return
}
func (r *GORMRepository) Get(ctx context.Context, id uuid.UUID) (*Application, error) {
	var v Application
	e := r.db.WithContext(ctx).Preload("Institution").Preload("Programme").Preload("Scholarship").First(&v, "id = ?", id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrApplicationNotFound
	}
	return &v, e
}
func (r *GORMRepository) GetDetail(ctx context.Context, id uuid.UUID) (*ApplicationDetail, error) {
	a, e := r.Get(ctx, id)
	if e != nil {
		return nil, e
	}
	d := &ApplicationDetail{Application: a}
	db := r.db.WithContext(ctx)
	for _, q := range []struct {
		dest  any
		order string
	}{{&d.Requirements, "sort_order NULLS LAST, created_at"}, {&d.Deadlines, "deadline_at NULLS LAST"}, {&d.Funding, "created_at"}, {&d.Contacts, "created_at"}, {&d.Supervisors, "created_at"}, {&d.URLs, "created_at"}, {&d.Tasks, "sort_order NULLS LAST, created_at"}} {
		if e = db.Where("application_id = ?", id).Order(q.order).Find(q.dest).Error; e != nil {
			return nil, e
		}
	}
	var run ResearchRun
	e = db.Where("application_id = ?", id).Order("created_at DESC").First(&run).Error
	if e == nil {
		d.LatestResearchRun = &run
	} else if !errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, e
	}
	return d, nil
}
func (r *GORMRepository) Update(ctx context.Context, a *Application, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Omit(clause.Associations).Save(a).Error; e != nil {
			return e
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}
func (r *GORMRepository) Delete(ctx context.Context, a *Application) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, m := range []any{&ApplicationRequirement{}, &ApplicationDeadline{}, &ApplicationFunding{}, &ApplicationContact{}, &ApplicationSupervisor{}, &ApplicationURL{}, &ApplicationTask{}} {
			if e := tx.Where("application_id = ?", a.ID).Delete(m).Error; e != nil {
				return e
			}
		}
		if e := tx.Delete(a).Error; e != nil {
			return e
		}
		return tx.Where("id = ?", a.ApplicantProfileID).Delete(&profile.ApplicantProfile{}).Error
	})
}

func getOwned[T any](db *gorm.DB, ctx context.Context, appID, id uuid.UUID, notFound error) (*T, error) {
	var v T
	e := db.WithContext(ctx).Where("application_id = ?", appID).First(&v, "id = ?", id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, notFound
	}
	return &v, e
}
func listOwned[T any](db *gorm.DB, ctx context.Context, appID uuid.UUID, order string) (v []T, e error) {
	e = db.WithContext(ctx).Where("application_id = ?", appID).Order(order).Find(&v).Error
	return
}
func createOne[T any](db *gorm.DB, ctx context.Context, v *T) error {
	return db.WithContext(ctx).Create(v).Error
}
func updateOne[T any](db *gorm.DB, ctx context.Context, v *T, a *profile.AuditLog) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Save(v).Error; e != nil {
			return e
		}
		if a != nil {
			return tx.Create(a).Error
		}
		return nil
	})
}
func deleteOne[T any](db *gorm.DB, ctx context.Context, v *T) error {
	return db.WithContext(ctx).Delete(v).Error
}

func (r *GORMRepository) ListRequirements(c context.Context, a uuid.UUID) ([]ApplicationRequirement, error) {
	return listOwned[ApplicationRequirement](r.db, c, a, "sort_order NULLS LAST, created_at")
}
func (r *GORMRepository) GetRequirement(c context.Context, a, id uuid.UUID) (*ApplicationRequirement, error) {
	return getOwned[ApplicationRequirement](r.db, c, a, id, ErrRequirementNotFound)
}
func (r *GORMRepository) CreateRequirement(c context.Context, v *ApplicationRequirement, audit *profile.AuditLog) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(v).Error; e != nil {
			return e
		}
		audit.EntityID = &v.ID
		return tx.Create(audit).Error
	})
}
func (r *GORMRepository) UpdateRequirement(c context.Context, v *ApplicationRequirement, a *profile.AuditLog) error {
	return updateOne(r.db, c, v, a)
}
func (r *GORMRepository) DeleteRequirement(c context.Context, v *ApplicationRequirement) error {
	return deleteOne(r.db, c, v)
}
func (r *GORMRepository) ListDeadlines(c context.Context, a uuid.UUID) ([]ApplicationDeadline, error) {
	return listOwned[ApplicationDeadline](r.db, c, a, "deadline_at NULLS LAST")
}
func (r *GORMRepository) GetDeadline(c context.Context, a, id uuid.UUID) (*ApplicationDeadline, error) {
	return getOwned[ApplicationDeadline](r.db, c, a, id, ErrDeadlineNotFound)
}
func (r *GORMRepository) CreateDeadline(c context.Context, v *ApplicationDeadline, audit *profile.AuditLog) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(v).Error; e != nil {
			return e
		}
		audit.EntityID = &v.ID
		return tx.Create(audit).Error
	})
}
func (r *GORMRepository) UpdateDeadline(c context.Context, v *ApplicationDeadline, a *profile.AuditLog) error {
	return updateOne(r.db, c, v, a)
}
func (r *GORMRepository) DeleteDeadline(c context.Context, v *ApplicationDeadline) error {
	return deleteOne(r.db, c, v)
}
func (r *GORMRepository) UpcomingDeadlines(c context.Context, from, to time.Time, appID, userID *uuid.UUID, status *Status) (v []ApplicationDeadline, e error) {
	q := r.db.WithContext(c).Where("deadline_at >= ? AND deadline_at <= ? AND date_precision IN ?", from, to, []string{"exact", "day"})
	if userID != nil {
		q = q.Where("application_id IN (SELECT id FROM applications WHERE user_id = ? AND deleted_at IS NULL)", *userID)
	}
	if appID != nil {
		q = q.Where("application_id = ?", *appID)
	}
	if status != nil {
		q = q.Where("application_id IN (?)", r.db.Model(&Application{}).Select("id").Where("status = ?", *status))
	}
	e = q.Order("deadline_at").Find(&v).Error
	return
}
func (r *GORMRepository) ListFunding(c context.Context, a uuid.UUID) ([]ApplicationFunding, error) {
	return listOwned[ApplicationFunding](r.db, c, a, "created_at")
}
func (r *GORMRepository) GetFunding(c context.Context, a, id uuid.UUID) (*ApplicationFunding, error) {
	return getOwned[ApplicationFunding](r.db, c, a, id, ErrFundingNotFound)
}
func (r *GORMRepository) CreateFunding(c context.Context, v *ApplicationFunding) error {
	return createOne(r.db, c, v)
}
func (r *GORMRepository) UpdateFunding(c context.Context, v *ApplicationFunding) error {
	return updateOne(r.db, c, v, nil)
}
func (r *GORMRepository) DeleteFunding(c context.Context, v *ApplicationFunding) error {
	return deleteOne(r.db, c, v)
}
func (r *GORMRepository) ListContacts(c context.Context, a uuid.UUID) ([]ApplicationContact, error) {
	return listOwned[ApplicationContact](r.db, c, a, "created_at")
}
func (r *GORMRepository) GetContact(c context.Context, a, id uuid.UUID) (*ApplicationContact, error) {
	return getOwned[ApplicationContact](r.db, c, a, id, ErrContactNotFound)
}
func (r *GORMRepository) CreateContact(c context.Context, v *ApplicationContact) error {
	return createOne(r.db, c, v)
}
func (r *GORMRepository) UpdateContact(c context.Context, v *ApplicationContact) error {
	return updateOne(r.db, c, v, nil)
}
func (r *GORMRepository) DeleteContact(c context.Context, v *ApplicationContact) error {
	return deleteOne(r.db, c, v)
}
func (r *GORMRepository) ListSupervisors(c context.Context, a uuid.UUID) ([]ApplicationSupervisor, error) {
	return listOwned[ApplicationSupervisor](r.db, c, a, "created_at")
}
func (r *GORMRepository) GetSupervisor(c context.Context, a, id uuid.UUID) (*ApplicationSupervisor, error) {
	return getOwned[ApplicationSupervisor](r.db, c, a, id, ErrSupervisorNotFound)
}
func (r *GORMRepository) CreateSupervisor(c context.Context, v *ApplicationSupervisor) error {
	return createOne(r.db, c, v)
}
func (r *GORMRepository) UpdateSupervisor(c context.Context, v *ApplicationSupervisor) error {
	return updateOne(r.db, c, v, nil)
}
func (r *GORMRepository) DeleteSupervisor(c context.Context, v *ApplicationSupervisor) error {
	return deleteOne(r.db, c, v)
}
func (r *GORMRepository) ListURLs(c context.Context, a uuid.UUID) ([]ApplicationURL, error) {
	return listOwned[ApplicationURL](r.db, c, a, "created_at")
}
func (r *GORMRepository) GetURL(c context.Context, a, id uuid.UUID) (*ApplicationURL, error) {
	return getOwned[ApplicationURL](r.db, c, a, id, ErrURLNotFound)
}
func (r *GORMRepository) CreateURL(c context.Context, v *ApplicationURL) error {
	return createOne(r.db, c, v)
}
func (r *GORMRepository) DeleteURL(c context.Context, v *ApplicationURL) error {
	return deleteOne(r.db, c, v)
}
func (r *GORMRepository) ListTasks(c context.Context, a uuid.UUID) ([]ApplicationTask, error) {
	return listOwned[ApplicationTask](r.db, c, a, "sort_order NULLS LAST, created_at")
}
func (r *GORMRepository) GetTask(c context.Context, a, id uuid.UUID) (*ApplicationTask, error) {
	return getOwned[ApplicationTask](r.db, c, a, id, ErrTaskNotFound)
}
func (r *GORMRepository) CreateTask(c context.Context, v *ApplicationTask) error {
	return createOne(r.db, c, v)
}
func (r *GORMRepository) UpdateTask(c context.Context, v *ApplicationTask, a *profile.AuditLog) error {
	return updateOne(r.db, c, v, a)
}
func (r *GORMRepository) DeleteTask(c context.Context, v *ApplicationTask) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Model(&ApplicationTask{}).Where("parent_task_id = ?", v.ID).Update("parent_task_id", nil).Error; e != nil {
			return e
		}
		return tx.Delete(v).Error
	})
}
func (r *GORMRepository) ListEvidence(c context.Context, requirementID uuid.UUID) (v []RequirementEvidence, e error) {
	e = r.db.WithContext(c).Where("requirement_id = ?", requirementID).Find(&v).Error
	return
}
func (r *GORMRepository) CreateEvidence(c context.Context, v *RequirementEvidence, a *profile.AuditLog) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(v).Error; e != nil {
			return e
		}
		a.EntityID = &v.ID
		return tx.Create(a).Error
	})
}
func (r *GORMRepository) Dashboard(c context.Context, userID *uuid.UUID) (*DashboardData, error) {
	d := &DashboardData{}
	db := r.db.WithContext(c)
	applications := db.Preload("Institution").Preload("Programme").Preload("Scholarship").Where("status NOT IN ?", []Status{StatusAccepted, StatusRejected, StatusWithdrawn, StatusExpired})
	if userID != nil {
		applications = applications.Where("user_id = ?", *userID)
	}
	if e := applications.Order("updated_at DESC").Limit(20).Find(&d.Applications).Error; e != nil {
		return nil, e
	}
	now := time.Now().UTC()
	deadlineQuery := db.Where("deadline_at >= ? AND deadline_at <= ? AND date_precision IN ?", now, now.AddDate(0, 0, 30), []string{"exact", "day"})
	taskQuery := db.Where("due_at >= ? AND due_at <= ? AND status NOT IN ?", now, now.AddDate(0, 0, 14), []string{"done", "cancelled"})
	ownerClause, ownerArgs := "", []any{}
	if userID != nil {
		ownerClause, ownerArgs = "application_id IN (SELECT id FROM applications WHERE user_id = ? AND deleted_at IS NULL)", []any{*userID}
		deadlineQuery, taskQuery = deadlineQuery.Where(ownerClause, ownerArgs...), taskQuery.Where(ownerClause, ownerArgs...)
	}
	if e := deadlineQuery.Order("deadline_at").Limit(20).Find(&d.Deadlines).Error; e != nil {
		return nil, e
	}
	if e := taskQuery.Order("due_at").Limit(20).Find(&d.Tasks).Error; e != nil {
		return nil, e
	}
	findingQuery := db.Model(&ResearchFinding{})
	requirementQuery := db.Model(&ApplicationRequirement{})
	if userID != nil {
		findingQuery = findingQuery.Where(ownerClause, ownerArgs...)
		requirementQuery = requirementQuery.Where(ownerClause, ownerArgs...)
	}
	findingQuery.Where("review_status = ?", "pending").Count(&d.PendingResearch)
	findingQuery.Where("verification_status = ? AND review_status = ?", "conflicting", "pending").Count(&d.ResearchConflicts)
	requirementQuery.Where("is_mandatory AND status NOT IN ?", []string{"satisfied", "waived", "not_applicable"}).Count(&d.IncompleteRequirements)
	return d, nil
}
