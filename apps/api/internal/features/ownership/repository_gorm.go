package ownership

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

type ownerRow struct {
	UserID uuid.UUID `gorm:"column:user_id"`
}

func (r *GORMRepository) owner(ctx context.Context, query string, id uuid.UUID) (uuid.UUID, error) {
	var row ownerRow
	result := r.db.WithContext(ctx).Raw(query, id).Scan(&row)
	if result.Error != nil {
		return uuid.Nil, result.Error
	}
	if result.RowsAffected == 0 || row.UserID == uuid.Nil {
		return uuid.Nil, ErrResourceNotFound
	}
	return row.UserID, nil
}

func (r *GORMRepository) ProfileOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, "SELECT user_id FROM applicant_profiles WHERE id = ? AND deleted_at IS NULL", id)
}

func (r *GORMRepository) ApplicationOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, "SELECT user_id FROM applications WHERE id = ? AND deleted_at IS NULL", id)
}

func (r *GORMRepository) ApplicationProposalOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, "SELECT user_id FROM application_proposals WHERE id = ?", id)
}

func (r *GORMRepository) QuestionnaireOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, `SELECT a.user_id
		FROM application_questionnaires q
		JOIN applications a ON a.id = q.application_id
		WHERE q.id = ? AND q.deleted_at IS NULL AND a.deleted_at IS NULL`, id)
}

func (r *GORMRepository) QuestionOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, `SELECT a.user_id
		FROM application_questions q
		JOIN application_questionnaires aq ON aq.id = q.questionnaire_id
		JOIN applications a ON a.id = aq.application_id
		WHERE q.id = ? AND q.deleted_at IS NULL AND aq.deleted_at IS NULL AND a.deleted_at IS NULL`, id)
}

func (r *GORMRepository) InformationRequestOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, "SELECT user_id FROM information_requests WHERE id = ?", id)
}

func (r *GORMRepository) ResearchTaskOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return r.owner(ctx, "SELECT user_id FROM research_tasks WHERE id = ? AND deleted_at IS NULL", id)
}
