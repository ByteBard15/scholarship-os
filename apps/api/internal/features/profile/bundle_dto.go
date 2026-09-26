package profile

import (
	"time"

	"github.com/google/uuid"
)

const ProfileBundleSchemaVersion = "1.0"

// ProfileBundle is the portable JSON representation of profile facts. It
// deliberately excludes documents, imports, evidence, snapshots, and audit
// records because those records are not portable profile content.
type ProfileBundle struct {
	SchemaVersion string            `json:"schemaVersion"`
	ExportedAt    time.Time         `json:"exportedAt"`
	Profile       ProfileBundleData `json:"profile"`
}

type ImportProfileBundleRequest struct {
	SchemaVersion string            `json:"schemaVersion" validate:"required"`
	ExportedAt    *time.Time        `json:"exportedAt,omitempty"`
	Profile       ProfileBundleData `json:"profile" validate:"required"`
}

type ProfileBundleData struct {
	Name              string               `json:"name" validate:"required"`
	ProfileType       ProfileType          `json:"profileType" validate:"required"`
	ParentProfileID   *uuid.UUID           `json:"parentProfileId,omitempty"`
	Headline          *string              `json:"headline,omitempty"`
	Summary           *string              `json:"summary,omitempty"`
	IsDefault         bool                 `json:"isDefault"`
	PersonalInfo      *PersonalInfoRequest `json:"personalInfo,omitempty"`
	Education         []SectionRequest     `json:"education"`
	Employment        []SectionRequest     `json:"employment"`
	Projects          []SectionRequest     `json:"projects"`
	Publications      []SectionRequest     `json:"publications"`
	Articles          []SectionRequest     `json:"articles"`
	Skills            []SectionRequest     `json:"skills"`
	ResearchInterests []SectionRequest     `json:"researchInterests"`
	CareerGoals       []SectionRequest     `json:"careerGoals"`
	Certifications    []SectionRequest     `json:"certifications"`
	Awards            []SectionRequest     `json:"awards"`
	Volunteering      []SectionRequest     `json:"volunteering"`
}
