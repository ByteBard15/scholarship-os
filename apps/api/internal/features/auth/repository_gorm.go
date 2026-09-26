package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMRepository struct{ db *gorm.DB }

func NewGORMRepository(db *gorm.DB) *GORMRepository { return &GORMRepository{db: db} }

func (r *GORMRepository) CreateAudit(ctx context.Context, value *profile.AuditLog) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *GORMRepository) CreateUser(ctx context.Context, value *user.User, master *profile.ApplicantProfile, audit *profile.AuditLog) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		if master != nil {
			master.UserID = value.ID
			if err := tx.Create(master).Error; err != nil {
				return err
			}
		}
		if audit != nil {
			audit.UserID = &value.ID
			return tx.Create(audit).Error
		}
		return nil
	})
	if err != nil && isDuplicate(err) {
		return ErrEmailAlreadyExists
	}
	return err
}

func (r *GORMRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var value user.User
	err := r.db.WithContext(ctx).Where("LOWER(BTRIM(email)) = ?", email).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrNotFound
	}
	return &value, err
}

func (r *GORMRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var value user.User
	err := r.db.WithContext(ctx).First(&value, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrNotFound
	}
	return &value, err
}

func (r *GORMRepository) ListUsers(ctx context.Context) ([]user.User, error) {
	var values []user.User
	return values, r.db.WithContext(ctx).Order("created_at DESC").Find(&values).Error
}

func (r *GORMRepository) UpdateUser(ctx context.Context, value *user.User, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(value).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) CreateSession(ctx context.Context, session *UserSession, value *user.User, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}
		if err := tx.Model(value).Update("last_login_at", value.LastLoginAt).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) GetSessionByHash(ctx context.Context, hash string) (*UserSession, error) {
	var value UserSession
	err := r.db.WithContext(ctx).Preload("User").Where("token_hash = ?", hash).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	return &value, err
}

func (r *GORMRepository) TouchSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).Where("id = ?", id).Update("last_used_at", time.Now().UTC()).Error
}

func (r *GORMRepository) RevokeSession(ctx context.Context, id uuid.UUID, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&UserSession{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) RevokeUserSessions(ctx context.Context, userID uuid.UUID, except *uuid.UUID, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&UserSession{}).Where("user_id = ? AND revoked_at IS NULL", userID)
		if except != nil {
			query = query.Where("id <> ?", *except)
		}
		if err := query.Update("revoked_at", time.Now().UTC()).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) UpdatePasswordAndRevokeSessions(ctx context.Context, value *user.User, except *uuid.UUID, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(value).Update("password_hash", value.PasswordHash).Error; err != nil {
			return err
		}
		query := tx.Model(&UserSession{}).Where("user_id = ? AND revoked_at IS NULL", value.ID)
		if except != nil {
			query = query.Where("id <> ?", *except)
		}
		if err := query.Update("revoked_at", time.Now().UTC()).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) CreateAgentCredential(ctx context.Context, value *AgentCredential, scopes []AgentCredentialScope, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		for i := range scopes {
			scopes[i].AgentCredentialID = value.ID
		}
		if err := tx.Create(&scopes).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (r *GORMRepository) GetAgentCredentialByHash(ctx context.Context, hash string) (*AgentCredential, error) {
	var value AgentCredential
	err := r.db.WithContext(ctx).Preload("User").Preload("Scopes").Where("key_hash = ?", hash).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	return &value, err
}

func (r *GORMRepository) GetAgentCredential(ctx context.Context, id uuid.UUID) (*AgentCredential, error) {
	var value AgentCredential
	err := r.db.WithContext(ctx).Preload("Scopes").First(&value, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCredentialNotFound
	}
	return &value, err
}

func (r *GORMRepository) ListAgentCredentials(ctx context.Context) ([]AgentCredential, error) {
	var values []AgentCredential
	return values, r.db.WithContext(ctx).Preload("Scopes").Order("created_at DESC").Find(&values).Error
}

func (r *GORMRepository) TouchAgentCredential(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&AgentCredential{}).Where("id = ?", id).Update("last_used_at", time.Now().UTC()).Error
}

func (r *GORMRepository) RevokeAgentCredential(ctx context.Context, value *AgentCredential, audit *profile.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		value.IsActive, value.RevokedAt = false, &now
		if err := tx.Save(value).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func isDuplicate(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "duplicate") || strings.Contains(text, "unique constraint")
}
