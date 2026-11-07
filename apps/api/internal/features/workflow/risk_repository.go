package workflow

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *GORMRepository) ListRiskAssessments(ctx context.Context, id uuid.UUID) ([]ProposalRiskAssessment, error) {
	rows := make([]ProposalRiskAssessment, 0)
	err := r.db.WithContext(ctx).Preload("Variables", func(tx *gorm.DB) *gorm.DB { return tx.Order("variable_name") }).Where("proposal_id = ?", id).Order("computed_at DESC, id DESC").Limit(50).Find(&rows).Error
	return rows, err
}
func (r *GORMRepository) CreateRiskAssessment(ctx context.Context, v *ProposalRiskAssessment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(v).Error
	})
}
