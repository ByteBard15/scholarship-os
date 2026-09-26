package profile

import (
	"context"
	"encoding/json"
	"errors"
	"math"

	"github.com/google/uuid"
)

func (s *Service) CreateSnapshot(ctx context.Context, profileID uuid.UUID, req CreateSnapshotRequest) (*ProfileSnapshot, error) {
	if s.workflow == nil {
		return nil, errors.New("profile workflow repository is not configured")
	}
	var created *ProfileSnapshot
	err := s.workflow.WithinTransaction(ctx, func(profiles Repository, sections SectionRepository, workflow WorkflowRepository) error {
		transactional := NewService(profiles, sections, s.users, WithWorkflowRepository(workflow))
		effective, err := transactional.ResolveEffectiveProfile(ctx, profileID)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(effective)
		if err != nil {
			return err
		}
		snapshot := &ProfileSnapshot{ProfileID: profileID, Snapshot: JSON(payload), CreatedBy: req.CreatedBy, Reason: req.Reason}
		if err = workflow.CreateSnapshot(ctx, snapshot); err != nil {
			return err
		}
		profile, err := profiles.GetByID(ctx, profileID)
		if err != nil {
			return err
		}
		entityType := "profile_snapshot"
		if err = workflow.CreateAudit(ctx, &AuditLog{UserID: &profile.UserID, ProfileID: &profileID, Action: "profile.snapshot.created", EntityType: &entityType, EntityID: &snapshot.ID}); err != nil {
			return err
		}
		created = snapshot
		return nil
	})
	return created, err
}
func (s *Service) ListSnapshots(ctx context.Context, profileID uuid.UUID) ([]ProfileSnapshot, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.ListSnapshots(ctx, profileID)
}
func (s *Service) GetSnapshot(ctx context.Context, profileID, id uuid.UUID) (*ProfileSnapshot, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.GetSnapshot(ctx, profileID, id)
}

func (s *Service) Completeness(ctx context.Context, profileID uuid.UUID) (*CompletenessResponse, error) {
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	sections := []CompletenessSectionResponse{}
	points := 0.0
	personalStatus := "missing"
	if effective.PersonalInfo != nil {
		personalStatus = "partial"
		points = .5
		if effective.PersonalInfo.Phone != nil && (effective.PersonalInfo.City != nil || effective.PersonalInfo.Country != nil) {
			personalStatus = "complete"
			points = 1
		}
	}
	sections = append(sections, CompletenessSectionResponse{Name: "personal_information", Status: personalStatus})
	for _, entry := range []struct {
		name  string
		count int
	}{{"education", len(effective.Education)}, {"employment", len(effective.Employment)}, {"projects", len(effective.Projects)}, {"skills", len(effective.Skills)}, {"research_interests", len(effective.ResearchInterests)}} {
		status := "missing"
		if entry.count > 0 {
			status = "complete"
			points++
		}
		sections = append(sections, CompletenessSectionResponse{Name: entry.name, Status: status})
	}
	return &CompletenessResponse{Score: int(math.Round(points / 6 * 100)), Sections: sections, Notice: "Completeness measures profile coverage only; it does not estimate applicant quality or scholarship likelihood."}, nil
}

func (s *Service) Compare(ctx context.Context, profileID, otherID uuid.UUID) (*ComparisonResponse, error) {
	base, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	other, err := s.profiles.GetByID(ctx, otherID)
	if err != nil {
		return nil, err
	}
	if base.UserID != other.UserID {
		return nil, ErrInvalidProfileParent
	}
	effective, err := s.ResolveEffectiveProfile(ctx, otherID)
	if err != nil {
		return nil, err
	}
	response := &ComparisonResponse{BaseProfile: profileResponse(base), ComparedProfile: profileResponse(other), InheritedEntities: []SectionResponse{}, HiddenEntities: []OverrideResponse{}, ModifiedFields: []OverrideResponse{}, AppendedEntities: []SectionResponse{}}
	for _, list := range [][]SectionResponse{effective.Education, effective.Employment, effective.Projects, effective.Publications, effective.Articles, effective.Skills, effective.ResearchInterests, effective.CareerGoals, effective.Certifications, effective.Awards, effective.Volunteering} {
		for _, item := range list {
			if item.ProfileID == otherID {
				response.AppendedEntities = append(response.AppendedEntities, item)
			} else {
				response.InheritedEntities = append(response.InheritedEntities, item)
			}
		}
	}
	lineage, err := s.Lineage(ctx, otherID)
	if err != nil {
		return nil, err
	}
	for _, profile := range lineage {
		if profile.ID == base.ID {
			continue
		}
		overrides, err := s.workflow.ListOverrides(ctx, profile.ID)
		if err != nil {
			return nil, err
		}
		for i := range overrides {
			item := overrideResponse(&overrides[i])
			if overrides[i].OverrideType == OverrideHide {
				response.HiddenEntities = append(response.HiddenEntities, item)
			} else {
				response.ModifiedFields = append(response.ModifiedFields, item)
			}
		}
	}
	return response, nil
}

func (s *Service) CreateEvidence(ctx context.Context, profileID uuid.UUID, req CreateEvidenceRequest) (*ProfileEvidence, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	entityType := normalizeEntityType(req.EntityType)
	if !(entityType == "profile" && req.EntityID == profileID) && findEffectiveEntity(effective, entityType, req.EntityID) == nil {
		return nil, ErrSectionNotFound
	}
	if req.DocumentID != nil {
		if _, err = s.workflow.GetDocument(ctx, profileID, *req.DocumentID); err != nil {
			return nil, err
		}
	}
	if req.ImportID != nil {
		if _, err = s.workflow.GetImport(ctx, profileID, *req.ImportID); err != nil {
			return nil, err
		}
	}
	v := &ProfileEvidence{ProfileID: profileID, EntityType: req.EntityType, EntityID: req.EntityID, FieldName: req.FieldName, EvidenceType: req.EvidenceType, DocumentID: req.DocumentID, ImportID: req.ImportID, SourceURL: req.SourceURL, SourceText: req.SourceText, Confidence: req.Confidence}
	if err := s.workflow.CreateEvidence(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *Service) ListEvidence(ctx context.Context, profileID uuid.UUID) ([]ProfileEvidence, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.ListEvidence(ctx, profileID)
}
