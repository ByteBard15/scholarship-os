package writing

import (
	"context"
	"sort"
	"strings"

	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateSample(ctx context.Context, request CreateSampleRequest, agentGenerated bool) (*SampleResponse, error) {
	userID, err := requestUserID(ctx, request.UserID)
	if err != nil {
		return nil, err
	}
	if request.ApplicationID != nil {
		valid, checkErr := s.repo.ApplicationBelongsToUser(ctx, *request.ApplicationID, userID)
		if checkErr != nil {
			return nil, checkErr
		}
		if !valid {
			return nil, ErrInvalidApplication
		}
	}
	sources := uniqueIDs(request.SourceSampleIDs)
	if agentGenerated && len(sources) == 0 {
		return nil, ErrInvalidSource
	}
	valid, err := s.repo.SamplesBelongToUser(ctx, userID, sources)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidSource
	}
	origin, status := request.Origin, request.Status
	if origin == "" {
		origin = OriginUser
	}
	if status == "" {
		status = StatusSource
	}
	var createdBy *string
	if agentGenerated {
		origin, status = OriginAgent, StatusSuggested
		value := "agent"
		createdBy = &value
	}
	sample := &WritingSample{
		ID: uuid.New(), UserID: userID, ApplicationID: request.ApplicationID,
		Title: strings.TrimSpace(request.Title), DocumentType: request.DocumentType,
		Content: strings.TrimSpace(request.Content), Origin: origin, Status: status,
		Description: trimmedPointer(request.Description), CreatedBy: createdBy,
	}
	if err := s.repo.CreateSample(ctx, sample, buildTags(userID, request.TagNames), sources); err != nil {
		return nil, err
	}
	response := sampleResponse(sample, sources)
	return &response, nil
}

func (s *Service) GetSample(ctx context.Context, id uuid.UUID) (*SampleResponse, error) {
	sample, err := s.repo.GetSample(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := principal.EnforceOwner(ctx, sample.UserID); err != nil {
		return nil, ErrSampleNotFound
	}
	sources, err := s.repo.ListSourceIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	response := sampleResponse(sample, sources)
	return &response, nil
}

func (s *Service) ListSamples(ctx context.Context, requestedUserID *uuid.UUID, filters SampleFilters) ([]SampleResponse, error) {
	userID, err := requestUserID(ctx, requestedUserID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListSamples(ctx, userID, filters)
	if err != nil {
		return nil, err
	}
	result := make([]SampleResponse, 0, len(items))
	for i := range items {
		sources, sourceErr := s.repo.ListSourceIDs(ctx, items[i].ID)
		if sourceErr != nil {
			return nil, sourceErr
		}
		result = append(result, sampleResponse(&items[i], sources))
	}
	return result, nil
}

func (s *Service) UpdateSample(ctx context.Context, id uuid.UUID, request UpdateSampleRequest) (*SampleResponse, error) {
	sample, err := s.repo.GetSample(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := principal.EnforceOwner(ctx, sample.UserID); err != nil {
		return nil, ErrSampleNotFound
	}
	if request.Title != nil {
		sample.Title = strings.TrimSpace(*request.Title)
	}
	if request.DocumentType != nil {
		sample.DocumentType = *request.DocumentType
	}
	if request.Content != nil {
		sample.Content = strings.TrimSpace(*request.Content)
	}
	if request.Status != nil {
		sample.Status = *request.Status
	}
	if request.Description != nil {
		sample.Description = trimmedPointer(request.Description)
	}
	var tags *[]WritingTag
	if request.TagNames != nil {
		values := buildTags(sample.UserID, *request.TagNames)
		tags = &values
	}
	if err := s.repo.UpdateSample(ctx, sample, tags); err != nil {
		return nil, err
	}
	sources, err := s.repo.ListSourceIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	response := sampleResponse(sample, sources)
	return &response, nil
}

func (s *Service) DeleteSample(ctx context.Context, id uuid.UUID) error {
	sample, err := s.repo.GetSample(ctx, id)
	if err != nil {
		return err
	}
	if err := principal.EnforceOwner(ctx, sample.UserID); err != nil {
		return ErrSampleNotFound
	}
	return s.repo.DeleteSample(ctx, sample)
}

func (s *Service) CreateTag(ctx context.Context, request CreateTagRequest) (*TagResponse, error) {
	userID, err := requestUserID(ctx, request.UserID)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(request.Name)
	tag := &WritingTag{ID: uuid.New(), UserID: userID, Name: name, NormalizedName: normalizeTag(name)}
	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	items, err := s.repo.ListTags(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if normalizeTag(items[i].Name) == tag.NormalizedName {
			return &items[i], nil
		}
	}
	return nil, ErrTagNotFound
}

func (s *Service) ListTags(ctx context.Context, requestedUserID *uuid.UUID) ([]TagResponse, error) {
	userID, err := requestUserID(ctx, requestedUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListTags(ctx, userID)
}

func (s *Service) DeleteTag(ctx context.Context, id uuid.UUID) error {
	tag, err := s.repo.GetTag(ctx, id)
	if err != nil {
		return err
	}
	if err := principal.EnforceOwner(ctx, tag.UserID); err != nil {
		return ErrTagNotFound
	}
	return s.repo.DeleteTag(ctx, tag)
}

func (s *Service) MatchSamples(ctx context.Context, request MatchSamplesRequest) ([]SampleMatchResponse, error) {
	userID, err := requestUserID(ctx, request.UserID)
	if err != nil {
		return nil, err
	}
	if request.ApplicationID != nil {
		valid, checkErr := s.repo.ApplicationBelongsToUser(ctx, *request.ApplicationID, userID)
		if checkErr != nil {
			return nil, checkErr
		}
		if !valid {
			return nil, ErrInvalidApplication
		}
	}
	filters := SampleFilters{DocumentType: request.DocumentType, Statuses: []string{StatusSource, StatusApproved}}
	samples, err := s.repo.ListSamples(ctx, userID, filters)
	if err != nil {
		return nil, err
	}
	requested := normalizedSet(request.Tags)
	if len(requested) == 0 {
		return nil, ErrTagsRequired
	}
	matches := make([]SampleMatchResponse, 0)
	for i := range samples {
		matched := make([]string, 0)
		for _, tag := range samples[i].Tags {
			if _, ok := requested[tag.NormalizedName]; ok {
				matched = append(matched, tag.Name)
			}
		}
		if len(matched) == 0 {
			continue
		}
		response := sampleResponse(&samples[i], nil)
		matches = append(matches, SampleMatchResponse{SampleResponse: response, MatchedTags: matched, Confidence: float64(len(matched)) / float64(len(requested))})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Confidence == matches[j].Confidence {
			return matches[i].UpdatedAt.After(matches[j].UpdatedAt)
		}
		return matches[i].Confidence > matches[j].Confidence
	})
	limit := request.Limit
	if limit == 0 {
		limit = 4
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

func requestUserID(ctx context.Context, requested *uuid.UUID) (uuid.UUID, error) {
	actor, ok := principal.PrincipalFromContext(ctx)
	if !ok {
		if requested != nil {
			return *requested, nil
		}
		return uuid.Nil, ErrUserRequired
	}
	if actor.IsSystem() {
		if requested == nil || *requested == uuid.Nil {
			return uuid.Nil, ErrUserRequired
		}
		return *requested, nil
	}
	if actor.UserID == nil {
		return uuid.Nil, ErrUserRequired
	}
	if requested != nil && *requested != *actor.UserID {
		return uuid.Nil, principal.ErrForbidden
	}
	return *actor.UserID, nil
}

func buildTags(userID uuid.UUID, names []string) []WritingTag {
	seen := map[string]struct{}{}
	tags := make([]WritingTag, 0, len(names))
	for _, value := range names {
		name := strings.TrimSpace(value)
		normalized := normalizeTag(name)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		tags = append(tags, WritingTag{ID: uuid.New(), UserID: userID, Name: name, NormalizedName: normalized})
	}
	return tags
}

func normalizedSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if normalized := normalizeTag(value); normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}

func normalizeTag(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func uniqueIDs(values []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func trimmedPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func sampleResponse(sample *WritingSample, sources []uuid.UUID) SampleResponse {
	tags := make([]TagResponse, 0, len(sample.Tags))
	for _, tag := range sample.Tags {
		tags = append(tags, TagResponse{ID: tag.ID, Name: tag.Name})
	}
	return SampleResponse{ID: sample.ID, UserID: sample.UserID, ApplicationID: sample.ApplicationID, Title: sample.Title, DocumentType: sample.DocumentType, Content: sample.Content, Origin: sample.Origin, Status: sample.Status, Description: sample.Description, CreatedBy: sample.CreatedBy, Tags: tags, SourceSampleIDs: sources, CreatedAt: sample.CreatedAt.UTC(), UpdatedAt: sample.UpdatedAt.UTC()}
}
