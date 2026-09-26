package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email          string    `gorm:"not null;uniqueIndex"`
	DisplayName    *string
	ExternalAuthID *string `gorm:"uniqueIndex"`
	PasswordHash   string  `gorm:"not null"`
	IsActive       bool    `gorm:"not null;default:true"`
	LastLoginAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (User) TableName() string { return "users" }
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
