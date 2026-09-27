package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/catalog"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct{ db *gorm.DB }

func NewGORMRepository(db *gorm.DB) *GORMRepository { return &GORMRepository{db: db} }

func (r *GORMRepository) CreateResearchTask(ctx context.Context, task *ResearchTask, links []ResearchTaskLink, newContexts []ResearchContext, contextIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for i := range links {
			links[i].ResearchTaskID = task.ID
		}
		if len(links) > 0 {
			if err := tx.Create(&links).Error; err != nil {
				return err
			}
		}
		if len(newContexts) > 0 {
			if err := tx.Create(&newContexts).Error; err != nil {
				return err
			}
			for i := range newContexts {
				contextIDs = append(contextIDs, newContexts[i].ID)
			}
		}
		return attachTaskContexts(tx, task.ID, contextIDs)
	})
}

func attachTaskContexts(tx *gorm.DB, taskID uuid.UUID, contextIDs []uuid.UUID) error {
	if len(contextIDs) == 0 {
		return nil
	}
	links := make([]ResearchTaskContext, 0, len(contextIDs))
	seen := make(map[uuid.UUID]struct{}, len(contextIDs))
	for _, contextID := range contextIDs {
		if _, exists := seen[contextID]; exists {
			continue
		}
		seen[contextID] = struct{}{}
		links = append(links, ResearchTaskContext{ResearchTaskID: taskID, ResearchContextID: contextID})
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error
}

func (r *GORMRepository) ListResearchTasks(ctx context.Context, filters ResearchTaskFilters) (items []ResearchTask, err error) {
	query := r.db.WithContext(ctx)
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.TaskType != nil {
		query = query.Where("task_type = ?", *filters.TaskType)
	}
	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}
	order := "created_at DESC"
	if filters.OldestFirst {
		order = "created_at ASC"
	}
	query = query.Order(order)
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	err = query.Find(&items).Error
	return
}

func (r *GORMRepository) ClaimResearchTask(ctx context.Context, id uuid.UUID, now time.Time) (task *ResearchTask, run *application.ResearchRun, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task = &ResearchTask{}
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(task, "id = ?", id).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return ErrResearchTaskNotFound
			}
			return queryErr
		}
		if task.Status != "queued" {
			return ErrInvalidResearchTaskState
		}
		task.Status, task.StartedAt, task.CompletedAt, task.FailedAt, task.FailureReason = "running", &now, nil, nil, nil
		run = &application.ResearchRun{Base: application.Base{ID: uuid.New()}, ResearchTaskID: &task.ID, ApplicationID: task.TargetApplicationID, Status: "running", Trigger: "agent", ResearchType: task.TaskType, StartedAt: &now}
		if createErr := tx.Create(run).Error; createErr != nil {
			return createErr
		}
		output := ResearchTaskOutput{ID: uuid.New(), ResearchTaskID: task.ID, OutputType: "research_run", EntityID: run.ID}
		if createErr := tx.Create(&output).Error; createErr != nil {
			return createErr
		}
		activity := AgentActivity{ID: uuid.New(), ResearchTaskID: &task.ID, ApplicationID: task.TargetApplicationID, ActivityType: "research_started", Summary: "Authenticated research agent claimed the queued task."}
		if createErr := tx.Create(&activity).Error; createErr != nil {
			return createErr
		}
		return tx.Save(task).Error
	})
	return
}

func (r *GORMRepository) CompleteResearchTask(ctx context.Context, taskID, runID uuid.UUID, now time.Time) (task *ResearchTask, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task = &ResearchTask{}
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(task, "id = ?", taskID).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return ErrResearchTaskNotFound
			}
			return queryErr
		}
		var run application.ResearchRun
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("research_task_id = ?", taskID).First(&run, "id = ?", runID).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return application.ErrResearchRunNotFound
			}
			return queryErr
		}
		if task.Status != "running" || run.Status != "running" {
			return ErrInvalidResearchTaskState
		}
		run.Status, run.CompletedAt = "review_required", &now
		task.Status, task.CompletedAt = "review_required", nil
		if saveErr := tx.Save(&run).Error; saveErr != nil {
			return saveErr
		}
		activity := AgentActivity{ID: uuid.New(), ResearchTaskID: &task.ID, ApplicationID: task.TargetApplicationID, PrefillRunID: nil, ActivityType: "research_completed", Summary: "Authenticated research agent finished submitting results; human review is required."}
		if createErr := tx.Create(&activity).Error; createErr != nil {
			return createErr
		}
		return tx.Save(task).Error
	})
	return
}

func (r *GORMRepository) GetResearchTask(ctx context.Context, id uuid.UUID) (*ResearchTask, error) {
	var task ResearchTask
	err := r.db.WithContext(ctx).First(&task, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResearchTaskNotFound
	}
	return &task, err
}

func (r *GORMRepository) UpdateResearchTask(ctx context.Context, task *ResearchTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *GORMRepository) DeleteResearchTask(ctx context.Context, task *ResearchTask) error {
	return r.db.WithContext(ctx).Delete(task).Error
}

func (r *GORMRepository) ResetResearchTaskResearch(ctx context.Context, task *ResearchTask) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		proposals := tx.Model(&ApplicationProposal{}).Select("id").Where("research_task_id = ?", task.ID)
		if err := tx.Where("application_proposal_id IN (?)", proposals).Delete(&ApplicationProposalSource{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id = ?", task.ID).Delete(&ApplicationProposal{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id = ?", task.ID).Delete(&ResearchTaskOutput{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id = ?", task.ID).Delete(&AgentActivity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("research_task_id = ?", task.ID).Delete(&application.ResearchRun{}).Error; err != nil {
			return err
		}
		activity := AgentActivity{ID: uuid.New(), ResearchTaskID: &task.ID, ApplicationID: task.TargetApplicationID, ActivityType: "research_restarted", Summary: "Previous research artifacts were cleared so the task can be researched again."}
		if err := tx.Create(&activity).Error; err != nil {
			return err
		}
		return tx.Save(task).Error
	})
}

func (r *GORMRepository) ListTaskLinks(ctx context.Context, taskID uuid.UUID) (items []ResearchTaskLink, err error) {
	err = r.db.WithContext(ctx).Where("research_task_id = ?", taskID).Order("created_at").Find(&items).Error
	return
}

func (r *GORMRepository) GetTaskLink(ctx context.Context, taskID, id uuid.UUID) (*ResearchTaskLink, error) {
	var link ResearchTaskLink
	err := r.db.WithContext(ctx).Where("research_task_id = ?", taskID).First(&link, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResearchTaskLinkNotFound
	}
	return &link, err
}

func (r *GORMRepository) CreateTaskLink(ctx context.Context, link *ResearchTaskLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *GORMRepository) UpdateTaskLink(ctx context.Context, link *ResearchTaskLink) error {
	return r.db.WithContext(ctx).Save(link).Error
}

func (r *GORMRepository) DeleteTaskLink(ctx context.Context, link *ResearchTaskLink) error {
	return r.db.WithContext(ctx).Delete(link).Error
}

func (r *GORMRepository) ListResearchContexts(ctx context.Context, userID *uuid.UUID) (items []ResearchContext, err error) {
	query := r.db.WithContext(ctx)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err = query.Order("updated_at DESC").Find(&items).Error
	return
}

func (r *GORMRepository) GetResearchContext(ctx context.Context, id uuid.UUID) (*ResearchContext, error) {
	var item ResearchContext
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResearchContextNotFound
	}
	return &item, err
}

func (r *GORMRepository) CreateResearchContext(ctx context.Context, item *ResearchContext) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GORMRepository) UpdateResearchContext(ctx context.Context, item *ResearchContext) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GORMRepository) DeleteResearchContext(ctx context.Context, item *ResearchContext) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

func (r *GORMRepository) ListTaskContexts(ctx context.Context, taskID uuid.UUID) (items []ResearchContext, err error) {
	err = r.db.WithContext(ctx).
		Joins("JOIN research_task_contexts rtc ON rtc.research_context_id = research_contexts.id").
		Where("rtc.research_task_id = ?", taskID).
		Order("rtc.created_at ASC").Find(&items).Error
	return
}

func (r *GORMRepository) AttachTaskContexts(ctx context.Context, taskID uuid.UUID, contextIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return attachTaskContexts(tx, taskID, contextIDs)
	})
}

func (r *GORMRepository) DetachTaskContext(ctx context.Context, taskID, contextID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("research_task_id = ? AND research_context_id = ?", taskID, contextID).Delete(&ResearchTaskContext{}).Error
}

func (r *GORMRepository) ListTaskOutputs(ctx context.Context, taskID uuid.UUID) (items []ResearchTaskOutput, err error) {
	err = r.db.WithContext(ctx).Where("research_task_id = ?", taskID).Order("created_at").Find(&items).Error
	return
}

func (r *GORMRepository) ListTaskRuns(ctx context.Context, taskID uuid.UUID) (items []application.ResearchRun, err error) {
	err = r.db.WithContext(ctx).Where("research_task_id = ?", taskID).Order("created_at DESC").Find(&items).Error
	return
}

func (r *GORMRepository) ListRunSources(ctx context.Context, runID uuid.UUID) (items []application.ResearchSource, err error) {
	err = r.db.WithContext(ctx).Where("research_run_id = ?", runID).Order("created_at").Find(&items).Error
	return
}

func (r *GORMRepository) ListRunFindings(ctx context.Context, runID uuid.UUID) (items []application.ResearchFinding, err error) {
	err = r.db.WithContext(ctx).Where("research_run_id = ?", runID).Order("created_at").Find(&items).Error
	return
}

func (r *GORMRepository) PersistResearchTaskResult(ctx context.Context, task *ResearchTask, data ResearchPersistence) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(data.Run).Error; err != nil {
			return err
		}
		if len(data.Sources) > 0 {
			if err := tx.Create(&data.Sources).Error; err != nil {
				return err
			}
		}
		if len(data.Findings) > 0 {
			if err := tx.Create(&data.Findings).Error; err != nil {
				return err
			}
		}
		if len(data.Proposals) > 0 {
			if err := tx.Create(&data.Proposals).Error; err != nil {
				return err
			}
		}
		if len(data.ProposalSources) > 0 {
			if err := tx.Create(&data.ProposalSources).Error; err != nil {
				return err
			}
		}
		if len(data.FollowUps) > 0 {
			if err := tx.Create(&data.FollowUps).Error; err != nil {
				return err
			}
			for i := range data.FollowUps {
				data.Outputs = append(data.Outputs, ResearchTaskOutput{ResearchTaskID: task.ID, OutputType: "follow_up_task", EntityID: data.FollowUps[i].ID})
			}
		}
		if len(data.Outputs) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&data.Outputs).Error; err != nil {
				return err
			}
		}
		if len(data.Activities) > 0 {
			if err := tx.Create(&data.Activities).Error; err != nil {
				return err
			}
		}
		return tx.Save(task).Error
	})
}

func (r *GORMRepository) CreateAgentSource(ctx context.Context, value *application.ResearchSource, activity *AgentActivity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		return tx.Create(activity).Error
	})
}

func (r *GORMRepository) CreateAgentFinding(ctx context.Context, value *application.ResearchFinding, activity *AgentActivity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		return tx.Create(activity).Error
	})
}

func (r *GORMRepository) CreateAgentProposal(ctx context.Context, task *ResearchTask, value *ApplicationProposal, sources []ApplicationProposalSource, output *ResearchTaskOutput, activity *AgentActivity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		if len(sources) > 0 {
			if err := tx.Create(&sources).Error; err != nil {
				return err
			}
		}
		if output != nil {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(output).Error; err != nil {
				return err
			}
		}
		if activity != nil {
			if err := tx.Create(activity).Error; err != nil {
				return err
			}
		}
		return tx.Save(task).Error
	})
}

func (r *GORMRepository) FailResearchTask(ctx context.Context, task *ResearchTask, run *application.ResearchRun, cause error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		message := cause.Error()
		run.Status, run.ErrorMessage = "failed", &message
		if run.ID == uuid.Nil {
			if err := tx.Create(run).Error; err != nil {
				return err
			}
		} else if err := tx.Save(run).Error; err != nil {
			return err
		}
		return tx.Save(task).Error
	})
}

func (r *GORMRepository) ListProposals(ctx context.Context, userID *uuid.UUID, status *string) (items []ApplicationProposal, err error) {
	query := r.db.WithContext(ctx)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err = query.Order("CASE priority WHEN 'highest' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END").Order("rank ASC NULLS LAST").Order("created_at ASC").Find(&items).Error
	return
}

func (r *GORMRepository) GetProposal(ctx context.Context, id uuid.UUID) (*ApplicationProposal, error) {
	var proposal ApplicationProposal
	err := r.db.WithContext(ctx).First(&proposal, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrApplicationProposalNotFound
	}
	return &proposal, err
}

func (r *GORMRepository) ListProposalSources(ctx context.Context, proposalID uuid.UUID) ([]ProposalSourceRecord, error) {
	var links []ApplicationProposalSource
	if err := r.db.WithContext(ctx).Where("application_proposal_id = ?", proposalID).Find(&links).Error; err != nil {
		return nil, err
	}
	items := make([]ProposalSourceRecord, 0, len(links))
	for _, link := range links {
		var source application.ResearchSource
		if err := r.db.WithContext(ctx).First(&source, "id = ?", link.ResearchSourceID).Error; err != nil {
			return nil, err
		}
		items = append(items, ProposalSourceRecord{Link: link, Source: source})
	}
	return items, nil
}

func (r *GORMRepository) UpdateProposal(ctx context.Context, proposal *ApplicationProposal, task *ResearchTask, activity AgentActivity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(proposal).Error; err != nil {
			return err
		}
		if task != nil {
			if err := tx.Save(task).Error; err != nil {
				return err
			}
		}
		return tx.Create(&activity).Error
	})
}

func (r *GORMRepository) ApproveProposalReview(ctx context.Context, proposal *ApplicationProposal, task *ResearchTask, parent *profile.ApplicantProfile, now time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked ApplicationProposal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", proposal.ID).Error; err != nil {
			return err
		}
		if locked.Status != "pending" {
			return ErrInvalidProposalState
		}
		locked.Status, locked.ReviewedAt, locked.ParentProfileID = "approved", &now, &parent.ID
		if err := tx.Save(&locked).Error; err != nil {
			return err
		}
		task.Status, task.CompletedAt = "review_required", nil
		if err := tx.Save(task).Error; err != nil {
			return err
		}
		activity := AgentActivity{ResearchTaskID: &task.ID, ActivityType: "proposal_approved", Summary: "Human approved the application proposal for agent processing."}
		if err := tx.Create(&activity).Error; err != nil {
			return err
		}
		*proposal = locked
		return nil
	})
}

func (r *GORMRepository) CreateApplicationFromProposal(ctx context.Context, proposal *ApplicationProposal, task *ResearchTask, parent *profile.ApplicantProfile) (*application.Application, error) {
	var created application.Application
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked ApplicationProposal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", proposal.ID).Error; err != nil {
			return err
		}
		if locked.Status != "approved" || locked.ApplicationID != nil || locked.ParentProfileID == nil || *locked.ParentProfileID != parent.ID {
			return ErrInvalidProposalState
		}
		institutionID, err := resolveInstitution(tx, &locked)
		if err != nil {
			return err
		}
		programmeID, err := resolveProgramme(tx, &locked, institutionID)
		if err != nil {
			return err
		}
		scholarshipID, err := resolveScholarship(tx, &locked, institutionID)
		if err != nil {
			return err
		}
		applicationProfile := profile.ApplicantProfile{Base: profile.Base{ID: uuid.New()}, UserID: locked.UserID, Name: locked.Name + " Application Profile", ProfileType: profile.ProfileTypeApplication, ParentProfileID: &parent.ID}
		if err = tx.Create(&applicationProfile).Error; err != nil {
			return err
		}
		created = application.Application{Base: application.Base{ID: uuid.New()}, UserID: locked.UserID, ApplicantProfileID: applicationProfile.ID, InstitutionID: institutionID, ProgrammeID: programmeID, ScholarshipID: scholarshipID, Name: locked.Name, Intake: locked.Intake, IntakeYear: locked.IntakeYear, Country: locked.Country, Status: application.StatusResearchComplete, ResearchStatus: application.ResearchComplete}
		if err = tx.Create(&created).Error; err != nil {
			return err
		}
		tasks := proposalDefaultTasks(created.ID)
		if err = tx.Create(&tasks).Error; err != nil {
			return err
		}
		if locked.ResearchRunID != nil {
			if err = tx.Model(&application.ResearchRun{}).Where("id = ?", *locked.ResearchRunID).Updates(map[string]any{"application_id": created.ID, "status": "complete"}).Error; err != nil {
				return err
			}
			if err = tx.Model(&application.ResearchSource{}).Where("research_run_id = ?", *locked.ResearchRunID).Update("application_id", created.ID).Error; err != nil {
				return err
			}
			if err = tx.Model(&application.ResearchFinding{}).Where("research_run_id = ?", *locked.ResearchRunID).Update("application_id", created.ID).Error; err != nil {
				return err
			}
			if err = applyProposalFindings(tx, created.ID, *locked.ResearchRunID); err != nil {
				return err
			}
		}
		now := time.Now().UTC()
		locked.ApplicationID = &created.ID
		if err = tx.Save(&locked).Error; err != nil {
			return err
		}
		var remaining int64
		if err = tx.Model(&ApplicationProposal{}).Where("research_task_id = ? AND id <> ? AND (status = ? OR (status = ? AND application_id IS NULL))", task.ID, locked.ID, "pending", "approved").Count(&remaining).Error; err != nil {
			return err
		}
		if remaining == 0 {
			task.Status, task.CompletedAt = "completed", &now
		} else {
			task.Status, task.CompletedAt = "review_required", nil
		}
		if err = tx.Save(task).Error; err != nil {
			return err
		}
		outputs := []ResearchTaskOutput{{ResearchTaskID: task.ID, OutputType: "application", EntityID: created.ID}, {ResearchTaskID: task.ID, OutputType: "application_profile", EntityID: applicationProfile.ID}}
		if err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&outputs).Error; err != nil {
			return err
		}
		entityType := "application"
		audit := profile.AuditLog{UserID: &locked.UserID, ProfileID: &applicationProfile.ID, Action: "application.created_from_approved_proposal", EntityType: &entityType, EntityID: &created.ID}
		if err = tx.Create(&audit).Error; err != nil {
			return err
		}
		activity := AgentActivity{ResearchTaskID: &task.ID, ApplicationID: &created.ID, ActivityType: "application_created", Summary: "Authenticated agent created the Application and isolated Application Profile from an approved proposal."}
		if err := tx.Create(&activity).Error; err != nil {
			return err
		}
		*proposal = locked
		return nil
	})
	return &created, err
}

func resolveInstitution(tx *gorm.DB, proposal *ApplicationProposal) (*uuid.UUID, error) {
	if proposal.InstitutionID != nil {
		return proposal.InstitutionID, nil
	}
	if len(proposal.ProposedInstitution) == 0 {
		return nil, nil
	}
	var data ProposedInstitution
	if err := json.Unmarshal(proposal.ProposedInstitution, &data); err != nil {
		return nil, err
	}
	var existing catalog.Institution
	err := tx.Where("LOWER(name) = LOWER(?) AND LOWER(country) = LOWER(?)", data.Name, data.Country).First(&existing).Error
	if err == nil {
		return &existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	created := catalog.Institution{Name: data.Name, ShortName: data.ShortName, InstitutionType: data.InstitutionType, Country: data.Country, City: data.City, WebsiteURL: data.WebsiteURL}
	if err = tx.Create(&created).Error; err != nil {
		return nil, err
	}
	return &created.ID, nil
}

func resolveProgramme(tx *gorm.DB, proposal *ApplicationProposal, institutionID *uuid.UUID) (*uuid.UUID, error) {
	if proposal.ProgrammeID != nil {
		return proposal.ProgrammeID, nil
	}
	if institutionID == nil || len(proposal.ProposedProgramme) == 0 {
		return nil, nil
	}
	var data ProposedProgramme
	if err := json.Unmarshal(proposal.ProposedProgramme, &data); err != nil {
		return nil, err
	}
	var existing catalog.Programme
	err := tx.Where("institution_id = ? AND LOWER(name) = LOWER(?)", *institutionID, data.Name).First(&existing).Error
	if err == nil {
		return &existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	created := catalog.Programme{InstitutionID: *institutionID, Name: data.Name, DegreeLevel: data.DegreeLevel, FieldOfStudy: data.FieldOfStudy, Faculty: data.Faculty, Department: data.Department, DurationMonths: data.DurationMonths, Mode: data.Mode, Language: data.Language, ProgrammeURL: data.ProgrammeURL, Description: data.Description}
	if err = tx.Create(&created).Error; err != nil {
		return nil, err
	}
	return &created.ID, nil
}

func resolveScholarship(tx *gorm.DB, proposal *ApplicationProposal, institutionID *uuid.UUID) (*uuid.UUID, error) {
	if proposal.ScholarshipID != nil {
		return proposal.ScholarshipID, nil
	}
	if len(proposal.ProposedScholarship) == 0 {
		return nil, nil
	}
	var data ProposedScholarship
	if err := json.Unmarshal(proposal.ProposedScholarship, &data); err != nil {
		return nil, err
	}
	var existing catalog.Scholarship
	query := tx.Where("LOWER(name) = LOWER(?)", data.Name)
	if data.ProviderName != nil {
		query = query.Where("LOWER(provider_name) = LOWER(?)", *data.ProviderName)
	}
	err := query.First(&existing).Error
	if err == nil {
		return &existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	created := catalog.Scholarship{InstitutionID: institutionID, Name: data.Name, ProviderName: data.ProviderName, Description: data.Description, Country: data.Country, DegreeLevel: data.DegreeLevel, ScholarshipType: data.ScholarshipType, OfficialURL: data.OfficialURL, IsRecurring: data.IsRecurring}
	if err = tx.Create(&created).Error; err != nil {
		return nil, err
	}
	return &created.ID, nil
}

func proposalDefaultTasks(applicationID uuid.UUID) []application.ApplicationTask {
	definitions := []struct{ title, kind string }{{"Verify research findings", "research"}, {"Review required documents", "document"}, {"Complete questionnaires", "application"}, {"Resolve missing information", "application"}, {"Review and submit application", "submission"}}
	items := make([]application.ApplicationTask, len(definitions))
	for i, definition := range definitions {
		order, source, kind := i+1, "proposal", definition.kind
		items[i] = application.ApplicationTask{ApplicationID: applicationID, Title: definition.title, Status: "todo", SortOrder: &order, Source: &source, TaskType: &kind}
	}
	return items
}

func applyProposalFindings(tx *gorm.DB, applicationID, runID uuid.UUID) error {
	var findings []application.ResearchFinding
	if err := tx.Where("research_run_id = ? AND verification_status IN ?", runID, []string{"verified", "supported"}).Find(&findings).Error; err != nil {
		return err
	}
	for i := range findings {
		finding := &findings[i]
		switch finding.Category {
		case "requirement":
			var request application.RequirementRequest
			if json.Unmarshal(finding.Value, &request) == nil {
				mandatory := request.IsMandatory == nil || *request.IsMandatory
				status := "unknown"
				if request.Status != nil {
					status = *request.Status
				}
				record := application.ApplicationRequirement{ApplicationID: applicationID, Category: request.Category, Title: request.Title, Description: request.Description, IsMandatory: mandatory, Status: status, DueDate: request.DueDate, SourceID: finding.SourceID, SourceURL: request.SourceURL, EvidenceRequired: request.EvidenceRequired != nil && *request.EvidenceRequired, Notes: request.Notes, SortOrder: request.SortOrder}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}
		case "deadline":
			var request application.DeadlineRequest
			if json.Unmarshal(finding.Value, &request) == nil {
				record := application.ApplicationDeadline{ApplicationID: applicationID, DeadlineType: request.DeadlineType, Title: request.Title, DeadlineAt: request.DeadlineAt, Timezone: request.Timezone, DatePrecision: request.DatePrecision, RawDeadlineText: request.RawDeadlineText, IsHardDeadline: request.IsHardDeadline != nil && *request.IsHardDeadline, SourceID: finding.SourceID, SourceURL: request.SourceURL, VerifiedAt: request.VerifiedAt, Notes: request.Notes}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}
		case "funding":
			var request application.FundingRequest
			if json.Unmarshal(finding.Value, &request) == nil {
				record := application.ApplicationFunding{ApplicationID: applicationID, FundingType: request.FundingType, Currency: request.Currency, Amount: request.Amount, AmountPeriod: request.AmountPeriod, TuitionCoverage: request.TuitionCoverage, StipendAmount: request.StipendAmount, StipendPeriod: request.StipendPeriod, TravelCoverage: request.TravelCoverage, InsuranceCoverage: request.InsuranceCoverage, AccommodationCoverage: request.AccommodationCoverage, OtherBenefits: request.OtherBenefits, Conditions: request.Conditions, SourceID: finding.SourceID, SourceURL: request.SourceURL}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}
		case "url":
			var request application.URLRequest
			if json.Unmarshal(finding.Value, &request) == nil {
				record := application.ApplicationURL{ApplicationID: applicationID, URLType: request.URLType, Label: request.Label, URL: request.URL, IsOfficial: request.IsOfficial}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}
		case "contact":
			var request application.ContactRequest
			if json.Unmarshal(finding.Value, &request) == nil {
				record := application.ApplicationContact{ApplicationID: applicationID, Name: request.Name, Role: request.Role, Email: request.Email, Phone: request.Phone, Organization: request.Organization, ContactType: request.ContactType, URL: request.URL, Notes: request.Notes, SourceID: finding.SourceID}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}
		case "supervisor":
			var request application.SupervisorRequest
			if json.Unmarshal(finding.Value, &request) == nil {
				record := application.ApplicationSupervisor{ApplicationID: applicationID, Name: request.Name, Title: request.Title, Department: request.Department, Institution: request.Institution, Email: request.Email, ProfileURL: request.ProfileURL, ResearchAreas: request.ResearchAreas, ContactStatus: request.ContactStatus, Notes: request.Notes, SourceID: finding.SourceID}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}
		case "questionnaire":
			var candidate QuestionnaireCandidate
			if json.Unmarshal(finding.Value, &candidate) == nil {
				questionnaire := ApplicationQuestionnaire{ApplicationID: applicationID, Title: candidate.Title, Description: candidate.Description, QuestionnaireType: candidate.QuestionnaireType, Status: "not_started", SourceURL: candidate.SourceURL}
				if err := tx.Create(&questionnaire).Error; err != nil {
					return err
				}
				for _, request := range candidate.Questions {
					status := "unanswered"
					if request.Status != nil {
						status = *request.Status
					}
					question := ApplicationQuestion{QuestionnaireID: questionnaire.ID, Key: request.Key, Prompt: request.Prompt, HelpText: request.HelpText, QuestionType: request.QuestionType, IsRequired: request.IsRequired, WordLimit: request.WordLimit, CharacterLimit: request.CharacterLimit, SortOrder: request.SortOrder, Options: profile.JSON(request.Options), Status: status}
					if err := tx.Create(&question).Error; err != nil {
						return err
					}
				}
			}
		}
		finding.ReviewStatus = "accepted"
		if err := tx.Model(finding).Update("review_status", finding.ReviewStatus).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *GORMRepository) ListFields(ctx context.Context, applicationID uuid.UUID) (items []ApplicationField, err error) {
	err = r.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("key").Find(&items).Error
	return
}

func (r *GORMRepository) GetField(ctx context.Context, applicationID, id uuid.UUID) (*ApplicationField, error) {
	var item ApplicationField
	err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrApplicationFieldNotFound
	}
	return &item, err
}

func (r *GORMRepository) CreateField(ctx context.Context, item *ApplicationField) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *GORMRepository) UpdateField(ctx context.Context, item *ApplicationField) error {
	return r.db.WithContext(ctx).Save(item).Error
}
func (r *GORMRepository) DeleteField(ctx context.Context, item *ApplicationField) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

func (r *GORMRepository) ListQuestionnaires(ctx context.Context, applicationID uuid.UUID) (items []ApplicationQuestionnaire, err error) {
	err = r.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("created_at").Find(&items).Error
	return
}

func (r *GORMRepository) GetQuestionnaire(ctx context.Context, applicationID, id uuid.UUID) (*ApplicationQuestionnaire, error) {
	var item ApplicationQuestionnaire
	err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrQuestionnaireNotFound
	}
	return &item, err
}

func (r *GORMRepository) GetQuestionnaireByID(ctx context.Context, id uuid.UUID) (*ApplicationQuestionnaire, error) {
	var item ApplicationQuestionnaire
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrQuestionnaireNotFound
	}
	return &item, err
}

func (r *GORMRepository) CreateQuestionnaire(ctx context.Context, item *ApplicationQuestionnaire) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *GORMRepository) UpdateQuestionnaire(ctx context.Context, item *ApplicationQuestionnaire) error {
	return r.db.WithContext(ctx).Save(item).Error
}
func (r *GORMRepository) DeleteQuestionnaire(ctx context.Context, item *ApplicationQuestionnaire) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

func (r *GORMRepository) ListQuestions(ctx context.Context, questionnaireID uuid.UUID) (items []ApplicationQuestion, err error) {
	err = r.db.WithContext(ctx).Where("questionnaire_id = ?", questionnaireID).Order("sort_order NULLS LAST, created_at").Find(&items).Error
	return
}

func (r *GORMRepository) GetQuestion(ctx context.Context, questionnaireID, id uuid.UUID) (*ApplicationQuestion, error) {
	var item ApplicationQuestion
	err := r.db.WithContext(ctx).Where("questionnaire_id = ?", questionnaireID).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrQuestionNotFound
	}
	return &item, err
}

func (r *GORMRepository) GetQuestionByID(ctx context.Context, id uuid.UUID) (*ApplicationQuestion, error) {
	var item ApplicationQuestion
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrQuestionNotFound
	}
	return &item, err
}

func (r *GORMRepository) CreateQuestion(ctx context.Context, item *ApplicationQuestion) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *GORMRepository) UpdateQuestion(ctx context.Context, item *ApplicationQuestion) error {
	return r.db.WithContext(ctx).Save(item).Error
}
func (r *GORMRepository) DeleteQuestion(ctx context.Context, item *ApplicationQuestion) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

func (r *GORMRepository) ListAnswers(ctx context.Context, questionID uuid.UUID) (items []ApplicationAnswer, err error) {
	err = r.db.WithContext(ctx).Where("question_id = ?", questionID).Order("created_at DESC").Find(&items).Error
	return
}

func (r *GORMRepository) GetAnswer(ctx context.Context, questionID, id uuid.UUID) (*ApplicationAnswer, error) {
	var item ApplicationAnswer
	err := r.db.WithContext(ctx).Where("question_id = ?", questionID).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAnswerNotFound
	}
	return &item, err
}

func (r *GORMRepository) CreateAnswer(ctx context.Context, item *ApplicationAnswer) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *GORMRepository) UpdateAnswer(ctx context.Context, item *ApplicationAnswer) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GORMRepository) RefreshQuestionnaireStatus(ctx context.Context, questionnaireID uuid.UUID) error {
	var required, resolved int64
	db := r.db.WithContext(ctx)
	if err := db.Model(&ApplicationQuestion{}).Where("questionnaire_id = ? AND is_required AND deleted_at IS NULL", questionnaireID).Count(&required).Error; err != nil {
		return err
	}
	if err := db.Model(&ApplicationQuestion{}).Where("questionnaire_id = ? AND is_required AND status IN ? AND deleted_at IS NULL", questionnaireID, []string{"answered", "approved"}).Count(&resolved).Error; err != nil {
		return err
	}
	status := "in_progress"
	if required > 0 && required == resolved {
		status = "completed"
	}
	if required == 0 {
		status = "not_started"
	}
	return db.Model(&ApplicationQuestionnaire{}).Where("id = ?", questionnaireID).Update("status", status).Error
}

func (r *GORMRepository) ListInformationRequests(ctx context.Context, filters InformationRequestFilters) (items []InformationRequest, err error) {
	query := r.db.WithContext(ctx)
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.ApplicationID != nil {
		query = query.Where("application_id = ?", *filters.ApplicationID)
	}
	if filters.RequestType != nil {
		query = query.Where("request_type = ?", *filters.RequestType)
	}
	err = query.Order("created_at DESC").Find(&items).Error
	return
}

func (r *GORMRepository) GetInformationRequest(ctx context.Context, id uuid.UUID) (*InformationRequest, error) {
	var item InformationRequest
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInformationRequestNotFound
	}
	return &item, err
}

func (r *GORMRepository) ListInformationResponses(ctx context.Context, requestID uuid.UUID) (items []InformationRequestResponse, err error) {
	err = r.db.WithContext(ctx).Where("information_request_id = ?", requestID).Order("created_at").Find(&items).Error
	return
}

func (r *GORMRepository) FindOpenInformationRequest(ctx context.Context, candidate InformationRequest) (*InformationRequest, error) {
	query := r.db.WithContext(ctx).Where("user_id = ? AND request_type = ? AND status IN ?", candidate.UserID, candidate.RequestType, []string{"pending", "answered", "processing", "reopened"})
	if candidate.ApplicationID != nil {
		query = query.Where("application_id = ?", *candidate.ApplicationID)
	}
	if candidate.QuestionID != nil {
		query = query.Where("question_id = ?", *candidate.QuestionID)
	} else if candidate.ApplicationFieldID != nil {
		query = query.Where("application_field_id = ?", *candidate.ApplicationFieldID)
	} else {
		return nil, gorm.ErrRecordNotFound
	}
	var item InformationRequest
	err := query.First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInformationRequestNotFound
	}
	return &item, err
}

func (r *GORMRepository) CreateInformationRequest(ctx context.Context, item *InformationRequest) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GORMRepository) RespondInformationRequest(ctx context.Context, item *InformationRequest, response *InformationRequestResponse) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(response).Error; err != nil {
			return err
		}
		return tx.Save(item).Error
	})
}

func (r *GORMRepository) UpdateInformationRequest(ctx context.Context, item *InformationRequest) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GORMRepository) ProcessInformationRequest(ctx context.Context, item *InformationRequest) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item.Status = "processing"
		if err := tx.Save(item).Error; err != nil {
			return err
		}
		if item.QuestionID != nil {
			answerSource, sourceType, createdBy := "information_request", "information_request", "agent"
			answer := ApplicationAnswer{QuestionID: *item.QuestionID, Value: item.ResponseValue, DraftText: item.ResponseText, Status: "answered", AnswerSource: &answerSource, SourceEntityType: &sourceType, SourceEntityID: &item.ID, CreatedBy: &createdBy}
			if err := tx.Create(&answer).Error; err != nil {
				return err
			}
			if err := tx.Model(&ApplicationQuestion{}).Where("id = ?", *item.QuestionID).Update("status", "answered").Error; err != nil {
				return err
			}
		} else if item.ApplicationFieldID != nil {
			value := item.ResponseValue
			if len(value) == 0 && item.ResponseText != nil {
				encoded, _ := json.Marshal(*item.ResponseText)
				value = profile.JSON(encoded)
			}
			if err := tx.Model(&ApplicationField{}).Where("id = ?", *item.ApplicationFieldID).Updates(map[string]any{"value": value, "status": "filled", "source_type": "information_request", "source_entity_id": item.ID}).Error; err != nil {
				return err
			}
		} else {
			return ErrInformationTargetRequired
		}
		now, resolution := time.Now().UTC(), "information_request"
		item.Status, item.CompletedAt, item.ResolutionSource = "completed", &now, &resolution
		if err := tx.Save(item).Error; err != nil {
			return err
		}
		activity := AgentActivity{ApplicationID: item.ApplicationID, ResearchTaskID: item.ResearchTaskID, ActivityType: "information_processed", Summary: "User-provided information was applied to its controlled target."}
		return tx.Create(&activity).Error
	})
}

func (r *GORMRepository) LoadPrefillInput(ctx context.Context, applicationID uuid.UUID) (*PrefillData, error) {
	data := &PrefillData{}
	var app application.Application
	if err := r.db.WithContext(ctx).Preload("Institution").Preload("Programme").Preload("Scholarship").First(&app, "id = ?", applicationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, application.ErrApplicationNotFound
		}
		return nil, err
	}
	data.Application = &app
	db := r.db.WithContext(ctx)
	if err := db.Where("application_id = ?", applicationID).Find(&data.Requirements).Error; err != nil {
		return nil, err
	}
	if err := db.Where("application_id = ?", applicationID).Find(&data.Fields).Error; err != nil {
		return nil, err
	}
	var questionnaires []ApplicationQuestionnaire
	if err := db.Where("application_id = ?", applicationID).Find(&questionnaires).Error; err != nil {
		return nil, err
	}
	for _, questionnaire := range questionnaires {
		bundle := QuestionnaireBundle{Questionnaire: questionnaire, Answers: map[uuid.UUID][]ApplicationAnswer{}}
		if err := db.Where("questionnaire_id = ?", questionnaire.ID).Order("sort_order NULLS LAST, created_at").Find(&bundle.Questions).Error; err != nil {
			return nil, err
		}
		for _, question := range bundle.Questions {
			var answers []ApplicationAnswer
			if err := db.Where("question_id = ?", question.ID).Order("created_at DESC").Find(&answers).Error; err != nil {
				return nil, err
			}
			bundle.Answers[question.ID] = answers
		}
		data.Questionnaires = append(data.Questionnaires, bundle)
	}
	if err := db.Where("application_id = ? AND status = ?", applicationID, "completed").Find(&data.CompletedInfo).Error; err != nil {
		return nil, err
	}
	if err := db.Where("application_id = ? AND review_status IN ?", applicationID, []string{"accepted", "auto_accepted"}).Find(&data.Findings).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (r *GORMRepository) PersistPrefillResult(ctx context.Context, data PrefillPersistence) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(data.Run).Error; err != nil {
			return err
		}
		for i := range data.Fields {
			data.Fields[i].PrefillRunID = &data.Run.ID
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "application_id"}, {Name: "key"}}, DoNothing: true}).Create(&data.Fields[i]).Error; err != nil {
				return err
			}
		}
		for i := range data.Answers {
			data.Answers[i].PrefillRunID = &data.Run.ID
			if err := tx.Create(&data.Answers[i]).Error; err != nil {
				return err
			}
			if err := tx.Model(&ApplicationQuestion{}).Where("id = ?", data.Answers[i].QuestionID).Update("status", questionStatusForAnswer(data.Answers[i].Status)).Error; err != nil {
				return err
			}
		}
		for i := range data.InformationRequests {
			candidate := data.InformationRequests[i]
			query := tx.Where("user_id = ? AND application_id = ? AND request_type = ? AND status IN ?", candidate.UserID, candidate.ApplicationID, candidate.RequestType, []string{"pending", "answered", "processing", "reopened"})
			if candidate.QuestionID != nil {
				query = query.Where("question_id = ?", *candidate.QuestionID)
			} else if candidate.ApplicationFieldID != nil {
				query = query.Where("application_field_id = ?", *candidate.ApplicationFieldID)
			}
			var count int64
			if err := query.Model(&InformationRequest{}).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				if err := tx.Create(&candidate).Error; err != nil {
					return err
				}
				if candidate.QuestionID != nil {
					if err := tx.Model(&ApplicationQuestion{}).Where("id = ?", *candidate.QuestionID).Update("status", "needs_information").Error; err != nil {
						return err
					}
				}
			}
		}
		for i := range data.Tasks {
			var count int64
			if err := tx.Model(&application.ApplicationTask{}).Where("application_id = ? AND LOWER(title) = LOWER(?) AND deleted_at IS NULL", data.Tasks[i].ApplicationID, data.Tasks[i].Title).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				if err := tx.Create(&data.Tasks[i]).Error; err != nil {
					return err
				}
			}
		}
		for i := range data.Actions {
			data.Actions[i].PrefillRunID = data.Run.ID
		}
		if len(data.Actions) > 0 {
			if err := tx.Create(&data.Actions).Error; err != nil {
				return err
			}
		}
		now := time.Now().UTC()
		data.Run.Status, data.Run.CompletedAt = "review_required", &now
		if err := tx.Save(data.Run).Error; err != nil {
			return err
		}
		for i := range data.Activities {
			data.Activities[i].PrefillRunID = &data.Run.ID
		}
		if len(data.Activities) > 0 {
			return tx.Create(&data.Activities).Error
		}
		return nil
	})
}

func questionStatusForAnswer(answerStatus string) string {
	if answerStatus == "suggested" {
		return "needs_review"
	}
	if answerStatus == "approved" {
		return "approved"
	}
	return "answered"
}

func (r *GORMRepository) PreparationSummary(ctx context.Context, applicationID uuid.UUID) (PreparationSummary, error) {
	var summary PreparationSummary
	db := r.db.WithContext(ctx)
	var fieldsTotal, fieldsCompleted, questionnairesTotal, questionnairesCompleted, pendingInformation, requiredTotal, requiredAnswered int64
	if err := db.Model(&ApplicationField{}).Where("application_id = ?", applicationID).Count(&fieldsTotal).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&ApplicationField{}).Where("application_id = ? AND status IN ?", applicationID, []string{"filled", "verified"}).Count(&fieldsCompleted).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&ApplicationQuestionnaire{}).Where("application_id = ? AND deleted_at IS NULL", applicationID).Count(&questionnairesTotal).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&ApplicationQuestionnaire{}).Where("application_id = ? AND status = ? AND deleted_at IS NULL", applicationID, "completed").Count(&questionnairesCompleted).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&InformationRequest{}).Where("application_id = ? AND status IN ?", applicationID, []string{"pending", "answered", "processing", "reopened"}).Count(&pendingInformation).Error; err != nil {
		return summary, err
	}
	questionIDs := db.Model(&ApplicationQuestionnaire{}).Select("id").Where("application_id = ? AND deleted_at IS NULL", applicationID)
	if err := db.Model(&ApplicationQuestion{}).Where("questionnaire_id IN (?) AND is_required AND deleted_at IS NULL", questionIDs).Count(&requiredTotal).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&ApplicationQuestion{}).Where("questionnaire_id IN (?) AND is_required AND status IN ? AND deleted_at IS NULL", questionIDs, []string{"answered", "approved"}).Count(&requiredAnswered).Error; err != nil {
		return summary, err
	}
	summary.FieldsTotal, summary.FieldsCompleted = int(fieldsTotal), int(fieldsCompleted)
	summary.QuestionnairesTotal, summary.QuestionnairesCompleted = int(questionnairesTotal), int(questionnairesCompleted)
	summary.PendingInformation = int(pendingInformation)
	summary.RequiredQuestionsTotal, summary.RequiredQuestionsAnswered = int(requiredTotal), int(requiredAnswered)
	return summary, nil
}
