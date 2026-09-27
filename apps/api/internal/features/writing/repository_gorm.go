package writing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct{ db *gorm.DB }

func NewGORMRepository(db *gorm.DB) *GORMRepository { return &GORMRepository{db: db} }

func (r *GORMRepository) CreateSample(ctx context.Context, sample *WritingSample, tags []WritingTag, sourceIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sample).Error; err != nil {
			return err
		}
		resolved, err := resolveTags(tx, tags)
		if err != nil {
			return err
		}
		if err := replaceSampleTags(tx, sample.ID, resolved); err != nil {
			return err
		}
		if len(sourceIDs) > 0 {
			now := time.Now().UTC()
			relations := make([]WritingSampleRelation, 0, len(sourceIDs))
			for _, sourceID := range sourceIDs {
				relations = append(relations, WritingSampleRelation{ID: uuid.New(), SourceSampleID: sourceID, DerivedSampleID: sample.ID, RelationType: "derived_from", CreatedAt: now})
			}
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		sample.Tags = resolved
		return nil
	})
}

func (r *GORMRepository) GetSample(ctx context.Context, id uuid.UUID) (*WritingSample, error) {
	var sample WritingSample
	if err := r.db.WithContext(ctx).Preload("Tags").First(&sample, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSampleNotFound
		}
		return nil, err
	}
	return &sample, nil
}

func (r *GORMRepository) ListSamples(ctx context.Context, userID uuid.UUID, filters SampleFilters) ([]WritingSample, error) {
	query := r.db.WithContext(ctx).Preload("Tags").Where("user_id = ?", userID).Order("updated_at DESC")
	if filters.DocumentType != nil {
		query = query.Where("document_type = ?", *filters.DocumentType)
	}
	if filters.ApplicationID != nil {
		query = query.Where("application_id = ?", *filters.ApplicationID)
	}
	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}
	samples := make([]WritingSample, 0)
	return samples, query.Find(&samples).Error
}

func (r *GORMRepository) UpdateSample(ctx context.Context, sample *WritingSample, tags *[]WritingTag) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(sample).Error; err != nil {
			return err
		}
		if tags == nil {
			return nil
		}
		resolved, err := resolveTags(tx, *tags)
		if err != nil {
			return err
		}
		if err := replaceSampleTags(tx, sample.ID, resolved); err != nil {
			return err
		}
		sample.Tags = resolved
		return nil
	})
}

func (r *GORMRepository) DeleteSample(ctx context.Context, sample *WritingSample) error {
	return r.db.WithContext(ctx).Delete(sample).Error
}

func (r *GORMRepository) ListSourceIDs(ctx context.Context, derivedID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	return ids, r.db.WithContext(ctx).Model(&WritingSampleRelation{}).Where("derived_sample_id = ?", derivedID).Pluck("source_sample_id", &ids).Error
}

func (r *GORMRepository) CreateTag(ctx context.Context, tag *WritingTag) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "normalized_name"}}, DoNothing: true}).Create(tag).Error
}

func (r *GORMRepository) ListTags(ctx context.Context, userID uuid.UUID) ([]TagResponse, error) {
	tags := make([]TagResponse, 0)
	err := r.db.WithContext(ctx).Table("writing_tags AS t").
		Select("t.id, t.name, COUNT(s.id) AS sample_count").
		Joins("LEFT JOIN writing_sample_tags st ON st.writing_tag_id = t.id").
		Joins("LEFT JOIN writing_samples s ON s.id = st.writing_sample_id AND s.deleted_at IS NULL AND s.status <> 'archived'").
		Where("t.user_id = ?", userID).
		Group("t.id, t.name").Order("t.name ASC").Scan(&tags).Error
	return tags, err
}

func (r *GORMRepository) GetTag(ctx context.Context, id uuid.UUID) (*WritingTag, error) {
	var tag WritingTag
	if err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	return &tag, nil
}

func (r *GORMRepository) DeleteTag(ctx context.Context, tag *WritingTag) error {
	return r.db.WithContext(ctx).Delete(tag).Error
}

func (r *GORMRepository) ApplicationBelongsToUser(ctx context.Context, applicationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("applications").Where("id = ? AND user_id = ? AND deleted_at IS NULL", applicationID, userID).Count(&count).Error
	return count == 1, err
}

func (r *GORMRepository) SamplesBelongToUser(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&WritingSample{}).Where("user_id = ? AND id IN ?", userID, ids).Count(&count).Error
	return count == int64(len(ids)), err
}

func resolveTags(tx *gorm.DB, tags []WritingTag) ([]WritingTag, error) {
	resolved := make([]WritingTag, 0, len(tags))
	for i := range tags {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "normalized_name"}}, DoNothing: true}).Create(&tags[i]).Error; err != nil {
			return nil, err
		}
		var stored WritingTag
		if err := tx.Where("user_id = ? AND normalized_name = ?", tags[i].UserID, tags[i].NormalizedName).First(&stored).Error; err != nil {
			return nil, err
		}
		resolved = append(resolved, stored)
	}
	return resolved, nil
}

func replaceSampleTags(tx *gorm.DB, sampleID uuid.UUID, tags []WritingTag) error {
	if err := tx.Where("writing_sample_id = ?", sampleID).Delete(&WritingSampleTag{}).Error; err != nil {
		return err
	}
	if len(tags) == 0 {
		return nil
	}
	now := time.Now().UTC()
	links := make([]WritingSampleTag, 0, len(tags))
	for _, tag := range tags {
		links = append(links, WritingSampleTag{WritingSampleID: sampleID, WritingTagID: tag.ID, CreatedAt: now})
	}
	return tx.Create(&links).Error
}
