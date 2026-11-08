package workflow

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type deletionScope struct {
	Preview                     ResearchDeletionPreview
	TaskIDs, AppIDs, ProfileIDs []uuid.UUID
}

// Both preview and execution use this graph, including previously hidden descendants.
func loadDeletionScope(tx *gorm.DB, root *ResearchTask, lock bool) (deletionScope, error) {
	scope := deletionScope{Preview: ResearchDeletionPreview{Tasks: []DeletionItem{}, Proposals: []DeletionItem{}, Applications: []DeletionItem{}, ArchivedProfiles: []DeletionItem{}, Blockers: []string{}}}
	var ids []uuid.UUID
	if err := tx.Raw(`WITH RECURSIVE subtree AS (SELECT id FROM research_tasks WHERE id = ? UNION SELECT child.id FROM research_tasks child JOIN subtree parent ON child.parent_task_id = parent.id) SELECT id FROM subtree ORDER BY id`, root.ID).Scan(&ids).Error; err != nil {
		return scope, err
	}
	var tasks []ResearchTask
	q := tx.Unscoped().Where("id IN ?", ids).Order("id")
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Find(&tasks).Error; err != nil {
		return scope, err
	}
	if len(tasks) == 0 {
		return scope, ErrResearchTaskNotFound
	}
	var found bool
	for _, t := range tasks {
		if t.UserID != root.UserID {
			return scope, ErrResearchTaskNotFound
		}
		if t.ID == root.ID {
			found = true
			if t.DeletedAt.Valid {
				return scope, ErrResearchTaskNotFound
			}
		}
		scope.TaskIDs = append(scope.TaskIDs, t.ID)
		scope.Preview.Tasks = append(scope.Preview.Tasks, DeletionItem{t.ID, t.Title})
		if t.Status == "running" {
			scope.Preview.Blockers = append(scope.Preview.Blockers, "A research task is running. Finish or stop its worker before deleting.")
		}
	}
	if !found {
		return scope, ErrResearchTaskNotFound
	}
	var proposals []ApplicationProposal
	q = tx.Where("research_task_id IN ?", scope.TaskIDs).Order("id")
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Find(&proposals).Error; err != nil {
		return scope, err
	}
	appSet := map[uuid.UUID]bool{}
	for _, p := range proposals {
		if p.UserID != root.UserID {
			return scope, ErrResearchTaskNotFound
		}
		scope.Preview.Proposals = append(scope.Preview.Proposals, DeletionItem{p.ID, p.Name})
		if p.ApplicationID != nil {
			appSet[*p.ApplicationID] = true
		}
	}
	for id := range appSet {
		scope.AppIDs = append(scope.AppIDs, id)
	}
	var apps []application.Application
	if len(scope.AppIDs) > 0 {
		q = tx.Unscoped().Where("id IN ?", scope.AppIDs).Order("id")
		if lock {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.Find(&apps).Error; err != nil {
			return scope, err
		}
		for _, a := range apps {
			if a.UserID != root.UserID {
				return scope, ErrResearchTaskNotFound
			}
			scope.Preview.Applications = append(scope.Preview.Applications, DeletionItem{a.ID, a.Name})
		}
		var shared int64
		if err := tx.Model(&ApplicationProposal{}).Where("application_id IN ? AND research_task_id NOT IN ?", scope.AppIDs, scope.TaskIDs).Count(&shared).Error; err != nil {
			return scope, err
		}
		if shared > 0 {
			scope.Preview.Blockers = append(scope.Preview.Blockers, "An application is linked to proposals outside this subtree. Resolve that shared link before deleting.")
		}
	}
	// Archive isolated application profiles, retaining immutable writing provenance.
	for _, a := range apps {
		var p profile.ApplicantProfile
		if err := tx.Unscoped().First(&p, "id = ?", a.ApplicantProfileID).Error; err != nil {
			return scope, err
		}
		if p.UserID != root.UserID {
			return scope, ErrResearchTaskNotFound
		}
		if p.ProfileType != profile.ProfileTypeApplication || p.DeletedAt.Valid {
			continue
		}
		var used int64
		for _, query := range []*gorm.DB{
			tx.Unscoped().Model(&application.Application{}).Where("applicant_profile_id = ? AND id NOT IN ?", p.ID, scope.AppIDs),
			tx.Unscoped().Model(&profile.ApplicantProfile{}).Where("parent_profile_id = ?", p.ID),
			tx.Model(&ApplicationProposal{}).Where("parent_profile_id = ? AND research_task_id NOT IN ?", p.ID, scope.TaskIDs),
			tx.Unscoped().Model(&ResearchTask{}).Where("profile_id = ? AND id NOT IN ?", p.ID, scope.TaskIDs),
		} {
			var n int64
			if err := query.Count(&n).Error; err != nil {
				return scope, err
			}
			used += n
		}
		if used == 0 {
			scope.ProfileIDs = append(scope.ProfileIDs, p.ID)
			scope.Preview.ArchivedProfiles = append(scope.Preview.ArchivedProfiles, DeletionItem{p.ID, p.Name})
		}
	}
	var active int64
	if err := tx.Model(&application.ResearchRun{}).Where("status = ? AND (research_task_id IN ? OR application_id IN ?)", "running", scope.TaskIDs, scope.AppIDs).Count(&active).Error; err != nil {
		return scope, err
	}
	if active > 0 {
		scope.Preview.Blockers = append(scope.Preview.Blockers, "A linked research run is running.")
	}
	if err := tx.Model(&ApplicationPrefillRun{}).Where("status = ? AND (research_task_id IN ? OR application_id IN ?)", "running", scope.TaskIDs, scope.AppIDs).Count(&active).Error; err != nil {
		return scope, err
	}
	if active > 0 {
		scope.Preview.Blockers = append(scope.Preview.Blockers, "A linked prefill run is running.")
	}
	raw, err := json.Marshal(struct {
		Tasks     []ResearchTask
		Proposals []ApplicationProposal
		Apps      []application.Application
		Profiles  []DeletionItem
	}{tasks, proposals, apps, scope.Preview.ArchivedProfiles})
	if err != nil {
		return scope, err
	}
	sum := sha256.Sum256(raw)
	scope.Preview.ConfirmationToken = hex.EncodeToString(sum[:])
	return scope, nil
}
func (r *GORMRepository) PreviewResearchDeletion(ctx context.Context, task *ResearchTask) (*ResearchDeletionPreview, error) {
	var result ResearchDeletionPreview
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scope, err := loadDeletionScope(tx, task, false)
		result = scope.Preview
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return &result, err
}
func (r *GORMRepository) CascadeDeleteResearch(ctx context.Context, task *ResearchTask, token string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scope, err := loadDeletionScope(tx, task, true)
		if err != nil {
			return err
		}
		if token != scope.Preview.ConfirmationToken {
			return validationError("deletion scope changed; refresh the preview and confirm again")
		}
		if len(scope.Preview.Blockers) > 0 {
			return validationError(scope.Preview.Blockers[0])
		}
		// Runs must precede tasks: a run cannot lose its only owner under its CHECK.
		if err := tx.Unscoped().Where("research_task_id IN ? OR application_id IN ?", scope.TaskIDs, scope.AppIDs).Delete(&application.ResearchRun{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id IN ? OR application_id IN ?", scope.TaskIDs, scope.AppIDs).Delete(&ApplicationPrefillRun{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id IN ? OR application_id IN ?", scope.TaskIDs, scope.AppIDs).Delete(&InformationRequest{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id IN ? OR application_id IN ?", scope.TaskIDs, scope.AppIDs).Delete(&AgentActivity{}).Error; err != nil {
			return err
		}
		// PostgreSQL FK cascades delete application-owned children, even hidden rows.
		if len(scope.AppIDs) > 0 {
			if err := tx.Unscoped().Where("id IN ?", scope.AppIDs).Delete(&application.Application{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("id IN ?", scope.TaskIDs).Delete(&ResearchTask{}).Error; err != nil {
			return err
		}
		if len(scope.ProfileIDs) > 0 {
			if err := tx.Model(&profile.ApplicantProfile{}).Where("id IN ?", scope.ProfileIDs).Updates(map[string]any{"deleted_at": time.Now().UTC(), "updated_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		entity := "research_task"
		metadata, _ := json.Marshal(scope.Preview)
		return tx.Create(&profile.AuditLog{UserID: &task.UserID, Action: "research_task.cascade_deleted", EntityType: &entity, EntityID: &task.ID, Metadata: profile.JSON(metadata), CreatedAt: time.Now().UTC()}).Error
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
}
