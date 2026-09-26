package catalog

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type Institution struct {
	Base
	Name            string `gorm:"not null"`
	ShortName       *string
	InstitutionType *string
	Country         string `gorm:"not null"`
	City            *string
	WebsiteURL      *string
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type Programme struct {
	Base
	InstitutionID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Name           string    `gorm:"not null"`
	DegreeLevel    *string
	FieldOfStudy   *string
	Faculty        *string
	Department     *string
	DurationMonths *int
	Mode           *string
	Language       *string
	ProgrammeURL   *string
	Description    *string        `gorm:"type:text"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type Scholarship struct {
	Base
	InstitutionID   *uuid.UUID `gorm:"type:uuid;index"`
	Name            string     `gorm:"not null"`
	ProviderName    *string
	Description     *string `gorm:"type:text"`
	Country         *string
	DegreeLevel     *string
	ScholarshipType *string
	OfficialURL     *string
	IsRecurring     bool
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
