package workflow

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type RiskVariableRequest struct {
	VariableName  string     `json:"variableName" validate:"required,max=100"`
	Value         *float64   `json:"value" validate:"omitempty,gte=0,lte=100"`
	Rationale     string     `json:"rationale" validate:"required,max=4000"`
	SourceURL     *string    `json:"sourceUrl" validate:"omitempty,max=2000"`
	ObservedAt    time.Time  `json:"observedAt" validate:"required"`
	TimeVariable  bool       `json:"timeVariable"`
	TimePredicate *string    `json:"timePredicate" validate:"omitempty,max=2000"`
	ValidUntil    *time.Time `json:"validUntil"`
	Blocking      bool       `json:"blocking"`
}
type CreateRiskAssessmentRequest struct {
	ProfileID uuid.UUID             `json:"profileId" validate:"required"`
	Notes     string                `json:"notes" validate:"required,max=10000"`
	Variables []RiskVariableRequest `json:"variables" validate:"required,min=1,max=30,dive"`
}
type RiskVariableResponse struct {
	ID uuid.UUID `json:"id"`
	RiskVariableRequest
	Stale bool `json:"stale"`
}
type RiskAssessmentResponse struct {
	ID                uuid.UUID              `json:"id"`
	ProposalID        uuid.UUID              `json:"proposalId"`
	ProfileSnapshotID uuid.UUID              `json:"profileSnapshotId"`
	ProposalSnapshot  json.RawMessage        `json:"proposalSnapshot"`
	RiskFactor        *float64               `json:"riskFactor"`
	Methodology       string                 `json:"methodology"`
	Notes             string                 `json:"notes"`
	RecordedBy        *uuid.UUID             `json:"recordedBy,omitempty"`
	RecordedByType    string                 `json:"recordedByType"`
	ComputedAt        time.Time              `json:"computedAt"`
	Stale             bool                   `json:"stale"`
	UnknownCount      int                    `json:"unknownCount"`
	BlockingCount     int                    `json:"blockingCount"`
	Variables         []RiskVariableResponse `json:"variables"`
}

func riskResponse(v ProposalRiskAssessment, now time.Time) RiskAssessmentResponse {
	r := RiskAssessmentResponse{ID: v.ID, ProposalID: v.ProposalID, ProfileSnapshotID: v.ProfileSnapshotID, ProposalSnapshot: json.RawMessage(v.ProposalSnapshot), RiskFactor: v.RiskFactor, Methodology: v.Methodology, Notes: v.Notes, RecordedBy: v.RecordedBy, RecordedByType: v.RecordedByType, ComputedAt: v.ComputedAt.UTC(), Variables: make([]RiskVariableResponse, 0, len(v.Variables))}
	for _, f := range v.Variables {
		stale := f.ValidUntil != nil && !now.Before(*f.ValidUntil)
		if stale {
			r.Stale = true
		}
		if f.Value == nil {
			r.UnknownCount++
		}
		if f.Blocking {
			r.BlockingCount++
		}
		r.Variables = append(r.Variables, RiskVariableResponse{ID: f.ID, RiskVariableRequest: RiskVariableRequest{VariableName: f.VariableName, Value: f.Value, Rationale: f.Rationale, SourceURL: f.SourceURL, ObservedAt: f.ObservedAt.UTC(), TimeVariable: f.TimeVariable, TimePredicate: f.TimePredicate, ValidUntil: f.ValidUntil, Blocking: f.Blocking}, Stale: stale})
	}
	return r
}
