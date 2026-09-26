package auth

import (
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash  string    `gorm:"size:64;not null;uniqueIndex"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	User       user.User `gorm:"foreignKey:UserID"`
}

func (UserSession) TableName() string { return "user_sessions" }
func (s *UserSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type AgentCredential struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     *uuid.UUID
	Name       string `gorm:"not null"`
	KeyPrefix  string `gorm:"not null"`
	KeyHash    string `gorm:"size:64;not null;uniqueIndex"`
	IsActive   bool   `gorm:"not null;default:true"`
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	User       *user.User             `gorm:"foreignKey:UserID"`
	Scopes     []AgentCredentialScope `gorm:"foreignKey:AgentCredentialID"`
}

func (AgentCredential) TableName() string { return "agent_credentials" }
func (c *AgentCredential) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type AgentCredentialScope struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	AgentCredentialID uuid.UUID `gorm:"type:uuid;not null;index"`
	Scope             string    `gorm:"not null"`
	CreatedAt         time.Time
}

func (AgentCredentialScope) TableName() string { return "agent_credential_scopes" }
func (s *AgentCredentialScope) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
