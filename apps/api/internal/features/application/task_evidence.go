package application

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/byte/scholarship-os/apps/api/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskEvidence records an observation, not approval or proof of submission.
type TaskEvidence struct {
	Base
	ApplicationTaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Kind              string
	Label             string
	Value             string
	SourceURL         *string
	RecordedBy        *uuid.UUID `gorm:"type:uuid"`
	RecordedByType    string
}

func (TaskEvidence) TableName() string { return "application_task_evidence" }

type TaskEvidenceRequest struct {
	Kind      string  `json:"kind" validate:"required,oneof=link identifier document confirmation note"`
	Label     string  `json:"label" validate:"required,max=200"`
	Value     string  `json:"value" validate:"required,max=10000"`
	SourceURL *string `json:"sourceUrl" validate:"omitempty,max=2000"`
}
type TaskEvidenceResponse struct {
	ID             uuid.UUID  `json:"id"`
	TaskID         uuid.UUID  `json:"taskId"`
	Kind           string     `json:"kind"`
	Label          string     `json:"label"`
	Value          string     `json:"value"`
	SourceURL      *string    `json:"sourceUrl,omitempty"`
	RecordedBy     *uuid.UUID `json:"recordedBy,omitempty"`
	RecordedByType string     `json:"recordedByType"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func evidenceResponse(v TaskEvidence) TaskEvidenceResponse {
	return TaskEvidenceResponse{v.ID, v.ApplicationTaskID, v.Kind, v.Label, v.Value, v.SourceURL, v.RecordedBy, v.RecordedByType, v.CreatedAt.UTC()}
}
func validEvidenceURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "" && u.User == nil
}
func validateTaskEvidence(q TaskEvidenceRequest) error {
	switch q.Kind {
	case "link", "identifier", "document", "confirmation", "note":
	default:
		return newValidationError("unsupported evidence kind")
	}
	if strings.TrimSpace(q.Label) == "" || len(q.Label) > 200 || strings.TrimSpace(q.Value) == "" || len(q.Value) > 10000 {
		return newValidationError("evidence label and value are required within their length limits")
	}
	if q.Kind == "link" && !validEvidenceURL(q.Value) {
		return newValidationError("link evidence must be an HTTP(S) URL without credentials")
	}
	if q.SourceURL != nil && (len(*q.SourceURL) > 2000 || !validEvidenceURL(*q.SourceURL)) {
		return newValidationError("sourceUrl must be an HTTP(S) URL without credentials")
	}
	return nil
}
func (s *Service) taskEvidenceAccess(ctx context.Context, appID, taskID uuid.UUID) (*Application, error) {
	actor, err := principal.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if actor.Type == principal.PrincipalAgent && !actor.HasScope(principal.ScopeTasks) {
		return nil, principal.ErrAgentScopeRequired
	}
	app, err := s.Get(ctx, appID)
	if err != nil {
		return nil, err
	}
	if _, err = s.repo.GetTask(ctx, appID, taskID); err != nil {
		return nil, err
	}
	return app, nil
}
func (s *Service) ListTaskEvidence(ctx context.Context, appID, taskID uuid.UUID) ([]TaskEvidence, error) {
	if _, err := s.taskEvidenceAccess(ctx, appID, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListTaskEvidence(ctx, taskID)
}
func (s *Service) CreateTaskEvidence(ctx context.Context, appID, taskID uuid.UUID, q TaskEvidenceRequest) (*TaskEvidence, error) {
	app, err := s.taskEvidenceAccess(ctx, appID, taskID)
	if err != nil {
		return nil, err
	}
	if err = validateTaskEvidence(q); err != nil {
		return nil, err
	}
	actor, _ := principal.PrincipalFromContext(ctx)
	recordedBy := actor.UserID
	if actor.Type == principal.PrincipalAgent {
		recordedBy = actor.AgentID
	}
	v := &TaskEvidence{Base: Base{ID: uuid.New()}, ApplicationTaskID: taskID, Kind: q.Kind, Label: strings.TrimSpace(q.Label), Value: strings.TrimSpace(q.Value), SourceURL: q.SourceURL, RecordedBy: recordedBy, RecordedByType: string(actor.Type)}
	err = s.repo.CreateTaskEvidence(ctx, v, s.audit(app, "application.task.evidence.recorded", "application_task", taskID))
	return v, err
}
func (r *GORMRepository) ListTaskEvidence(ctx context.Context, taskID uuid.UUID) ([]TaskEvidence, error) {
	items := make([]TaskEvidence, 0)
	err := r.db.WithContext(ctx).Where("application_task_id = ?", taskID).Order("created_at, id").Find(&items).Error
	return items, err
}
func (r *GORMRepository) CreateTaskEvidence(ctx context.Context, v *TaskEvidence, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(v).Error; err != nil {
			return err
		}
		return tx.Create(audit).Error
	})
}
func (h *Handler) ListTaskEvidence(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.appID(w, r)
	if !ok {
		return
	}
	taskID, ok := h.id(w, r, "taskID", "INVALID_TASK_ID")
	if !ok {
		return
	}
	items, err := h.service.ListTaskEvidence(r.Context(), appID, taskID)
	if err != nil {
		h.err(w, r, err)
		return
	}
	result := make([]TaskEvidenceResponse, 0, len(items))
	for _, v := range items {
		result = append(result, evidenceResponse(v))
	}
	response.Collection(w, result, len(result))
}
func (h *Handler) CreateTaskEvidence(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.appID(w, r)
	if !ok {
		return
	}
	taskID, ok := h.id(w, r, "taskID", "INVALID_TASK_ID")
	if !ok {
		return
	}
	var q TaskEvidenceRequest
	if !h.body(w, r, &q, true) {
		return
	}
	v, err := h.service.CreateTaskEvidence(r.Context(), appID, taskID, q)
	if err != nil {
		h.err(w, r, err)
		return
	}
	response.Data(w, http.StatusCreated, evidenceResponse(*v))
}
