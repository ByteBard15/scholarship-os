package profile

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (r *GORMRepository) contentSafeDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Discard})
}

func (r *GORMRepository) WithinTransaction(ctx context.Context, fn func(Repository, SectionRepository, WorkflowRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewGORMRepository(tx)
		return fn(txRepo, txRepo, txRepo)
	})
}

func (r *GORMRepository) CreateOverride(ctx context.Context, v *ProfileOverride) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetOverride(ctx context.Context, pid, id uuid.UUID) (*ProfileOverride, error) {
	var v ProfileOverride
	err := r.db.WithContext(ctx).First(&v, "profile_id = ? AND id = ?", pid, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileOverrideNotFound
	}
	return &v, err
}
func (r *GORMRepository) ListOverrides(ctx context.Context, pid uuid.UUID) ([]ProfileOverride, error) {
	var v []ProfileOverride
	err := r.db.WithContext(ctx).Where("profile_id = ?", pid).Order("created_at ASC").Find(&v).Error
	return v, err
}
func (r *GORMRepository) UpdateOverride(ctx context.Context, v *ProfileOverride) error {
	return r.db.WithContext(ctx).Save(v).Error
}
func (r *GORMRepository) DeleteOverride(ctx context.Context, pid, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("profile_id = ? AND id = ?", pid, id).Delete(&ProfileOverride{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProfileOverrideNotFound
	}
	return nil
}

func (r *GORMRepository) CreateSnapshot(ctx context.Context, v *ProfileSnapshot) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(hashtext(?))", v.ProfileID.String()).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Model(&ProfileSnapshot{}).Where("profile_id = ?", v.ProfileID).Select("COALESCE(MAX(version), 0) + 1").Scan(&v.Version).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetSnapshot(ctx context.Context, pid, id uuid.UUID) (*ProfileSnapshot, error) {
	var v ProfileSnapshot
	err := r.db.WithContext(ctx).First(&v, "profile_id = ? AND id = ?", pid, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSnapshotNotFound
	}
	return &v, err
}
func (r *GORMRepository) ListSnapshots(ctx context.Context, pid uuid.UUID) ([]ProfileSnapshot, error) {
	var v []ProfileSnapshot
	err := r.db.WithContext(ctx).Where("profile_id = ?", pid).Order("version DESC").Find(&v).Error
	return v, err
}

func (r *GORMRepository) CreateDocument(ctx context.Context, v *ProfileDocument) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetDocument(ctx context.Context, pid, id uuid.UUID) (*ProfileDocument, error) {
	var v ProfileDocument
	err := r.db.WithContext(ctx).First(&v, "profile_id = ? AND id = ?", pid, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileDocumentNotFound
	}
	return &v, err
}
func (r *GORMRepository) ListDocuments(ctx context.Context, pid uuid.UUID) ([]ProfileDocument, error) {
	var v []ProfileDocument
	err := r.db.WithContext(ctx).Where("profile_id = ?", pid).Order("created_at DESC").Find(&v).Error
	return v, err
}
func (r *GORMRepository) DeleteDocument(ctx context.Context, pid, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("profile_id = ? AND id = ?", pid, id).Delete(&ProfileDocument{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProfileDocumentNotFound
	}
	return nil
}

func (r *GORMRepository) CreateImport(ctx context.Context, v *ProfileImport) error {
	return r.db.WithContext(ctx).Create(v).Error
}
func (r *GORMRepository) GetImport(ctx context.Context, pid, id uuid.UUID) (*ProfileImport, error) {
	var v ProfileImport
	err := r.db.WithContext(ctx).First(&v, "profile_id = ? AND id = ?", pid, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileImportNotFound
	}
	return &v, err
}
func (r *GORMRepository) ListImports(ctx context.Context, pid uuid.UUID) ([]ProfileImport, error) {
	var v []ProfileImport
	err := r.db.WithContext(ctx).Where("profile_id = ?", pid).Order("created_at DESC").Find(&v).Error
	return v, err
}
func (r *GORMRepository) UpdateImport(ctx context.Context, v *ProfileImport) error {
	return r.contentSafeDB(ctx).Save(v).Error
}
func (r *GORMRepository) CreateCandidates(ctx context.Context, v []ProfileImportCandidate) error {
	if len(v) == 0 {
		return nil
	}
	return r.contentSafeDB(ctx).Create(&v).Error
}
func (r *GORMRepository) GetCandidate(ctx context.Context, importID, id uuid.UUID) (*ProfileImportCandidate, error) {
	var v ProfileImportCandidate
	err := r.db.WithContext(ctx).First(&v, "import_id = ? AND id = ?", importID, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrImportCandidateNotFound
	}
	return &v, err
}
func (r *GORMRepository) ListCandidates(ctx context.Context, importID uuid.UUID) ([]ProfileImportCandidate, error) {
	var v []ProfileImportCandidate
	err := r.db.WithContext(ctx).Where("import_id = ?", importID).Order("created_at ASC").Find(&v).Error
	return v, err
}
func (r *GORMRepository) UpdateCandidate(ctx context.Context, v *ProfileImportCandidate) error {
	return r.contentSafeDB(ctx).Save(v).Error
}
func (r *GORMRepository) ApplyImport(ctx context.Context, imp *ProfileImport, entities []any, evidence []ProfileEvidence, audit *AuditLog) error {
	return r.contentSafeDB(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Create(entity).Error; err != nil {
				return err
			}
		}
		if len(evidence) > 0 {
			if err := tx.Create(&evidence).Error; err != nil {
				return err
			}
		}
		if err := tx.Save(imp).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) CreateEvidence(ctx context.Context, v *ProfileEvidence) error {
	return r.contentSafeDB(ctx).Create(v).Error
}
func (r *GORMRepository) ListEvidence(ctx context.Context, pid uuid.UUID) ([]ProfileEvidence, error) {
	var v []ProfileEvidence
	err := r.db.WithContext(ctx).Where("profile_id = ?", pid).Order("created_at DESC").Find(&v).Error
	return v, err
}
func (r *GORMRepository) CreateAudit(ctx context.Context, v *AuditLog) error {
	return r.db.WithContext(ctx).Create(v).Error
}
