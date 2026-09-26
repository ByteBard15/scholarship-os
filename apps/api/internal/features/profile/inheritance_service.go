package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) Lineage(ctx context.Context, profileID uuid.UUID) ([]ApplicantProfile, error) {
	current, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	lineage := []ApplicantProfile{*current}
	seen := map[uuid.UUID]bool{current.ID: true}
	for current.ParentProfileID != nil {
		if len(lineage) >= 3 {
			return nil, ErrProfileInheritanceCycle
		}
		parent, err := s.profiles.GetByID(ctx, *current.ParentProfileID)
		if err != nil {
			return nil, ErrInvalidProfileParent
		}
		if parent.UserID != current.UserID || seen[parent.ID] {
			return nil, ErrProfileInheritanceCycle
		}
		seen[parent.ID] = true
		lineage = append(lineage, *parent)
		current = parent
	}
	for left, right := 0, len(lineage)-1; left < right; left, right = left+1, right-1 {
		lineage[left], lineage[right] = lineage[right], lineage[left]
	}
	return lineage, nil
}

func (s *Service) ResolveEffectiveProfile(ctx context.Context, profileID uuid.UUID) (*EffectiveProfileResponse, error) {
	if s.workflow == nil {
		return nil, errors.New("profile workflow repository is not configured")
	}
	lineage, err := s.Lineage(ctx, profileID)
	if err != nil {
		return nil, err
	}
	var effective FullProfileResponse
	for index, item := range lineage {
		full, err := s.profiles.GetFull(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		current := fullResponse(full)
		if index == 0 {
			effective = current
		} else {
			mergeDirectContent(&effective, current)
			effective.ProfileResponse = current.ProfileResponse
		}
		overrides, err := s.workflow.ListOverrides(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		for i := range overrides {
			if err = applyOverride(&effective, &overrides[i], item.ID); err != nil {
				return nil, fmt.Errorf("resolve override %s: %w", overrides[i].ID, err)
			}
		}
	}
	target := lineage[len(lineage)-1]
	markInheritance(&effective, target.ID)
	items := make([]LineageItemResponse, 0, len(lineage))
	for _, item := range lineage {
		items = append(items, LineageItemResponse{ID: item.ID, Name: item.Name, ProfileType: item.ProfileType})
	}
	return &EffectiveProfileResponse{FullProfileResponse: effective, Lineage: items}, nil
}

func mergeDirectContent(target *FullProfileResponse, current FullProfileResponse) {
	if current.PersonalInfo != nil {
		target.PersonalInfo = current.PersonalInfo
	}
	target.Education = append(target.Education, current.Education...)
	target.Employment = append(target.Employment, current.Employment...)
	target.Projects = append(target.Projects, current.Projects...)
	target.Publications = append(target.Publications, current.Publications...)
	target.Articles = append(target.Articles, current.Articles...)
	target.Skills = append(target.Skills, current.Skills...)
	target.ResearchInterests = append(target.ResearchInterests, current.ResearchInterests...)
	target.CareerGoals = append(target.CareerGoals, current.CareerGoals...)
	target.Certifications = append(target.Certifications, current.Certifications...)
	target.Awards = append(target.Awards, current.Awards...)
	target.Volunteering = append(target.Volunteering, current.Volunteering...)
}
func markInheritance(v *FullProfileResponse, targetID uuid.UUID) {
	if v.PersonalInfo != nil && v.PersonalInfo.ProfileID != targetID {
		source := v.PersonalInfo.ProfileID
		v.PersonalInfo.Inherited = true
		v.PersonalInfo.InheritedFromProfileID = &source
	}
	for _, list := range []*[]SectionResponse{&v.Education, &v.Employment, &v.Projects, &v.Publications, &v.Articles, &v.Skills, &v.ResearchInterests, &v.CareerGoals, &v.Certifications, &v.Awards, &v.Volunteering} {
		for i := range *list {
			if (*list)[i].ProfileID != targetID {
				source := (*list)[i].ProfileID
				(*list)[i].Inherited = true
				(*list)[i].InheritedFromProfileID = &source
			}
		}
	}
}

func applyOverride(effective *FullProfileResponse, override *ProfileOverride, profileID uuid.UUID) error {
	entityType := normalizeEntityType(override.EntityType)
	if entityType == "profile" {
		return applyProfileOverride(&effective.ProfileResponse, override)
	}
	if entityType == "personal_info" {
		if effective.PersonalInfo == nil {
			return ErrSectionNotFound
		}
		if override.OverrideType == OverrideHide {
			effective.PersonalInfo = nil
			return nil
		}
		if err := applyJSONField(effective.PersonalInfo, override); err != nil {
			return err
		}
		effective.PersonalInfo.ModifiedInProfileID = &profileID
		return nil
	}
	list, ok := sectionList(effective, entityType)
	if !ok {
		return newValidationError("unsupported override entityType")
	}
	if override.EntityID == nil {
		return newValidationError("entityId is required for section overrides")
	}
	for i := range *list {
		if (*list)[i].ID != *override.EntityID {
			continue
		}
		if override.OverrideType == OverrideHide {
			*list = append((*list)[:i], (*list)[i+1:]...)
			return nil
		}
		if err := applyJSONField(&(*list)[i], override); err != nil {
			return err
		}
		(*list)[i].ModifiedInProfileID = &profileID
		return nil
	}
	return ErrSectionNotFound
}
func applyProfileOverride(profile *ProfileResponse, override *ProfileOverride) error {
	if override.OverrideType == OverrideHide {
		return newValidationError("profiles cannot be hidden")
	}
	switch override.FieldName {
	case "name":
		return decodeStringOverride(&profile.Name, override)
	case "headline":
		var value string
		if err := json.Unmarshal(override.Value, &value); err != nil {
			return err
		}
		if override.OverrideType == OverrideAppend && profile.Headline != nil {
			value = *profile.Headline + "\n" + value
		}
		profile.Headline = &value
	case "summary":
		var value string
		if err := json.Unmarshal(override.Value, &value); err != nil {
			return err
		}
		if override.OverrideType == OverrideAppend && profile.Summary != nil {
			value = *profile.Summary + "\n" + value
		}
		profile.Summary = &value
	default:
		return newValidationError("profile field is not overridable")
	}
	return nil
}
func decodeStringOverride(target *string, override *ProfileOverride) error {
	var value string
	if err := json.Unmarshal(override.Value, &value); err != nil {
		return err
	}
	if override.OverrideType == OverrideAppend && *target != "" {
		value = *target + "\n" + value
	}
	*target = value
	return nil
}
func applyJSONField(target any, override *ProfileOverride) error {
	if override.FieldName == "id" || override.FieldName == "profileId" || override.FieldName == "createdAt" || override.FieldName == "updatedAt" {
		return newValidationError("field cannot be overridden")
	}
	raw, err := json.Marshal(target)
	if err != nil {
		return err
	}
	var values map[string]any
	if err = json.Unmarshal(raw, &values); err != nil {
		return err
	}
	current, exists := values[override.FieldName]
	if !exists {
		return newValidationError("field does not exist on entity")
	}
	var next any
	if err = json.Unmarshal(override.Value, &next); err != nil {
		return newValidationError("override value must be valid JSON")
	}
	if override.OverrideType == OverrideAppend {
		left, leftOK := current.(string)
		right, rightOK := next.(string)
		if !leftOK || !rightOK {
			return newValidationError("append overrides require string fields")
		}
		next = strings.TrimSpace(left + "\n" + right)
	}
	values[override.FieldName] = next
	updated, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return json.Unmarshal(updated, target)
}
func normalizeEntityType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "education_history":
		return "education"
	case "employment_history":
		return "employment"
	case "project":
		return "projects"
	case "publication":
		return "publications"
	case "article":
		return "articles"
	case "skill":
		return "skills"
	case "research-interest", "research-interests", "researchinterests":
		return "research_interests"
	case "career-goal", "career-goals", "careergoals":
		return "career_goals"
	case "personalinfo":
		return "personal_info"
	case "certification":
		return "certifications"
	case "award":
		return "awards"
	case "volunteer":
		return "volunteering"
	}
	return value
}
func sectionList(v *FullProfileResponse, entityType string) (*[]SectionResponse, bool) {
	switch entityType {
	case "education":
		return &v.Education, true
	case "employment":
		return &v.Employment, true
	case "projects":
		return &v.Projects, true
	case "publications":
		return &v.Publications, true
	case "articles":
		return &v.Articles, true
	case "skills":
		return &v.Skills, true
	case "research_interests":
		return &v.ResearchInterests, true
	case "career_goals":
		return &v.CareerGoals, true
	case "certifications":
		return &v.Certifications, true
	case "awards":
		return &v.Awards, true
	case "volunteering":
		return &v.Volunteering, true
	}
	return nil, false
}

func (s *Service) CreateOverride(ctx context.Context, profileID uuid.UUID, req CreateOverrideRequest) (*ProfileOverride, error) {
	profile, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if profile.ProfileType == ProfileTypeMaster {
		return nil, newValidationError("master profiles contain canonical facts; create overrides on derived profiles")
	}
	if req.OverrideType != OverrideReplace && req.OverrideType != OverrideHide && req.OverrideType != OverrideAppend {
		return nil, newValidationError("overrideType must be replace, hide, or append")
	}
	if req.OverrideType != OverrideHide && len(req.Value) == 0 {
		return nil, newValidationError("value is required")
	}
	v := &ProfileOverride{ProfileID: profileID, EntityType: normalizeEntityType(req.EntityType), EntityID: req.EntityID, FieldName: req.FieldName, OverrideType: req.OverrideType, Value: JSON(req.Value), Reason: req.Reason, Source: req.Source}
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	probe := effective.FullProfileResponse
	if err = applyOverride(&probe, v, profileID); err != nil {
		return nil, err
	}
	if err = s.workflow.CreateOverride(ctx, v); err != nil {
		return nil, err
	}
	entityType := v.EntityType
	_ = s.workflow.CreateAudit(ctx, &AuditLog{UserID: &profile.UserID, ProfileID: &profileID, Action: "profile.override.created", EntityType: &entityType, EntityID: &v.ID})
	return v, nil
}
func (s *Service) ListOverrides(ctx context.Context, profileID uuid.UUID) ([]ProfileOverride, error) {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	return s.workflow.ListOverrides(ctx, profileID)
}
func (s *Service) UpdateOverride(ctx context.Context, profileID, id uuid.UUID, req UpdateOverrideRequest) (*ProfileOverride, error) {
	v, err := s.workflow.GetOverride(ctx, profileID, id)
	if err != nil {
		return nil, err
	}
	if req.Value != nil {
		v.Value = JSON(req.Value)
	}
	if req.Reason != nil {
		v.Reason = req.Reason
	}
	if req.Source != nil {
		v.Source = req.Source
	}
	if v.OverrideType != OverrideHide && len(v.Value) == 0 {
		return nil, newValidationError("value is required")
	}
	effective, err := s.ResolveEffectiveProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	probe := effective.FullProfileResponse
	if err = applyOverride(&probe, v, profileID); err != nil {
		return nil, err
	}
	if err = s.workflow.UpdateOverride(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *Service) DeleteOverride(ctx context.Context, profileID, id uuid.UUID) error {
	profile, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return err
	}
	if err = s.workflow.DeleteOverride(ctx, profileID, id); err != nil {
		return err
	}
	entityType := "profile_override"
	_ = s.workflow.CreateAudit(ctx, &AuditLog{UserID: &profile.UserID, ProfileID: &profileID, Action: "profile.override.deleted", EntityType: &entityType, EntityID: &id})
	return nil
}
