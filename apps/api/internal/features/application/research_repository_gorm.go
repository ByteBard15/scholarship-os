package application

import (
	"context"
	"errors"

	"github.com/example/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *GORMRepository) CreateRun(c context.Context, v *ResearchRun, app *Application) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(v).Error; e != nil {
			return e
		}
		return tx.Model(&Application{}).Where("id = ?", v.ApplicationID).Updates(map[string]any{"research_status": ResearchRunning, "status": app.Status}).Error
	})
}
func (r *GORMRepository) PersistResult(c context.Context, run *ResearchRun, app *Application, sources []ResearchSource, findings []ResearchFinding) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if len(sources) > 0 {
			if e := tx.Create(&sources).Error; e != nil {
				return e
			}
		}
		if len(findings) > 0 {
			if e := tx.Create(&findings).Error; e != nil {
				return e
			}
		}
		if e := tx.Save(run).Error; e != nil {
			return e
		}
		return tx.Model(&Application{}).Where("id = ?", app.ID).Updates(map[string]any{"research_status": app.ResearchStatus, "status": app.Status}).Error
	})
}
func (r *GORMRepository) FailRun(c context.Context, run *ResearchRun, app *Application, cause error) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		message := cause.Error()
		run.Status = "failed"
		run.ErrorMessage = &message
		app.ResearchStatus = ResearchFailed
		if e := tx.Save(run).Error; e != nil {
			return e
		}
		return tx.Model(&Application{}).Where("id = ?", app.ID).Update("research_status", ResearchFailed).Error
	})
}
func (r *GORMRepository) ListRuns(c context.Context, app uuid.UUID) (v []ResearchRun, e error) {
	e = r.db.WithContext(c).Where("application_id = ?", app).Order("created_at DESC").Find(&v).Error
	return
}
func (r *GORMRepository) GetRun(c context.Context, app, id uuid.UUID) (*ResearchRun, error) {
	return getOwned[ResearchRun](r.db, c, app, id, ErrResearchRunNotFound)
}
func (r *GORMRepository) ListSources(c context.Context, run uuid.UUID) (v []ResearchSource, e error) {
	e = r.db.WithContext(c).Where("research_run_id = ?", run).Order("created_at").Find(&v).Error
	return
}
func (r *GORMRepository) CreateSource(c context.Context, v *ResearchSource) error {
	return r.db.WithContext(c).Create(v).Error
}
func (r *GORMRepository) ListFindings(c context.Context, run uuid.UUID) (v []ResearchFinding, e error) {
	e = r.db.WithContext(c).Where("research_run_id = ?", run).Order("created_at").Find(&v).Error
	return
}
func (r *GORMRepository) CreateFinding(c context.Context, v *ResearchFinding) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(v).Error; e != nil {
			return e
		}
		if e := tx.Model(&ResearchRun{}).Where("id = ?", v.ResearchRunID).Update("status", "review_required").Error; e != nil {
			return e
		}
		return tx.Model(&Application{}).Where("id = ?", v.ApplicationID).Update("research_status", ResearchReviewRequired).Error
	})
}
func (r *GORMRepository) GetFinding(c context.Context, run, id uuid.UUID) (*ResearchFinding, error) {
	var v ResearchFinding
	e := r.db.WithContext(c).Where("research_run_id = ?", run).First(&v, "id = ?", id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrResearchFindingNotFound
	}
	return &v, e
}
func (r *GORMRepository) UpdateFinding(c context.Context, v *ResearchFinding) error {
	return r.db.WithContext(c).Model(v).Update("review_status", v.ReviewStatus).Error
}
func (r *GORMRepository) ApplyFindings(c context.Context, run *ResearchRun, app *Application, b AppliedResearch, audit *profile.AuditLog) error {
	return r.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		for _, items := range []any{b.Requirements, b.Deadlines, b.Funding, b.Contacts, b.Supervisors, b.URLs} {
			switch v := items.(type) {
			case []ApplicationRequirement:
				if len(v) > 0 {
					if e := tx.Create(&v).Error; e != nil {
						return e
					}
				}
			case []ApplicationDeadline:
				if len(v) > 0 {
					if e := tx.Create(&v).Error; e != nil {
						return e
					}
				}
			case []ApplicationFunding:
				if len(v) > 0 {
					if e := tx.Create(&v).Error; e != nil {
						return e
					}
				}
			case []ApplicationContact:
				if len(v) > 0 {
					if e := tx.Create(&v).Error; e != nil {
						return e
					}
				}
			case []ApplicationSupervisor:
				if len(v) > 0 {
					if e := tx.Create(&v).Error; e != nil {
						return e
					}
				}
			case []ApplicationURL:
				if len(v) > 0 {
					if e := tx.Create(&v).Error; e != nil {
						return e
					}
				}
			}
		}
		run.Status = "complete"
		app.ResearchStatus = ResearchComplete
		if app.Status == StatusResearching || app.Status == StatusDiscovered {
			app.Status = StatusResearchComplete
		}
		if e := tx.Save(run).Error; e != nil {
			return e
		}
		if e := tx.Model(&Application{}).Where("id = ?", app.ID).Updates(map[string]any{"research_status": app.ResearchStatus, "status": app.Status}).Error; e != nil {
			return e
		}
		return tx.Create(audit).Error
	})
}
func (r *GORMRepository) MarkStale(c context.Context, app *Application) error {
	return r.db.WithContext(c).Model(&Application{}).Where("id = ?", app.ID).Update("research_status", ResearchStale).Error
}
