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
