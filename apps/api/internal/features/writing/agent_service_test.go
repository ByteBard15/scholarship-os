package writing

import (
	"context"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	principal "github.com/byte/scholarship-os/apps/api/pkg/auth"
	"github.com/google/uuid"
)

type writingMemory struct {
	Repository
	owner         uuid.UUID
	items         map[uuid.UUID]*WritingSample
	sources       map[uuid.UUID][]uuid.UUID
	creates       int
	metadataOwner uuid.UUID
}

func (r *writingMemory) SamplesBelongToUser(_ context.Context, owner uuid.UUID, ids []uuid.UUID) (bool, error) {
	for _, id := range ids {
		if r.items[id] == nil || r.items[id].UserID != owner {
			return false, nil
		}
	}
	return true, nil
}
func (r *writingMemory) ApplicationBelongsToUser(_ context.Context, _ uuid.UUID, owner uuid.UUID) (bool, error) {
	return owner == r.owner, nil
}
func (r *writingMemory) CreateSample(_ context.Context, s *WritingSample, tags []WritingTag, ids []uuid.UUID) error {
	s.Tags = tags
	r.items[s.ID] = s
	r.sources[s.ID] = ids
	r.creates++
	return nil
}
func (r *writingMemory) GetSample(_ context.Context, id uuid.UUID) (*WritingSample, error) {
	s := r.items[id]
	if s == nil {
		return nil, ErrSampleNotFound
	}
	copy := *s
	return &copy, nil
}
func (r *writingMemory) ListSourceIDs(_ context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	return r.sources[id], nil
}
func (r *writingMemory) UpdateSuggestedSample(_ context.Context, s *WritingSample, _ *[]WritingTag) error {
	r.items[s.ID] = s
	return nil
}
func (r *writingMemory) ListSampleMetadata(_ context.Context, owner uuid.UUID, _ SampleMetadataQuery) ([]SampleMetadataResponse, error) {
	r.metadataOwner = owner
	return []SampleMetadataResponse{}, nil
}
func writingContext(owner uuid.UUID, scopes ...string) context.Context {
	m := map[string]struct{}{}
	for _, s := range scopes {
		m[s] = struct{}{}
	}
	return principal.WithPrincipal(context.Background(), principal.Principal{Type: principal.PrincipalAgent, UserID: &owner, Scopes: m})
}

type writingProfiles struct {
	owner, id, snapshot uuid.UUID
	snapshots           int
}

func (p *writingProfiles) Get(context.Context, uuid.UUID) (*profile.ApplicantProfile, error) {
	return &profile.ApplicantProfile{Base: profile.Base{ID: p.id}, UserID: p.owner}, nil
}
func (p *writingProfiles) CreateSnapshot(context.Context, uuid.UUID, profile.CreateSnapshotRequest) (*profile.ProfileSnapshot, error) {
	p.snapshots++
	return &profile.ProfileSnapshot{ID: p.snapshot, ProfileID: p.id}, nil
}
func TestGeneratedWritingPreservesFactualSourcesAndReview(t *testing.T) {
	owner, pid := uuid.New(), uuid.New()
	r := &writingMemory{owner: owner, items: map[uuid.UUID]*WritingSample{}, sources: map[uuid.UUID][]uuid.UUID{}}
	profiles := &writingProfiles{owner: owner, id: pid, snapshot: uuid.New()}
	s := NewService(r, profiles)
	ctx := writingContext(owner, principal.ScopeWriting, principal.ScopeProfiles)
	req := CreateSampleRequest{Title: "Supervisor enquiry", DocumentType: TypeEmail, Content: "Supported biography", Origin: OriginUser, Status: StatusApproved, SourceProfileID: &pid, TagNames: []string{" Calgary ", "calgary"}}
	result, err := s.CreateSample(ctx, req, false)
	if err != nil || result.Status != StatusSuggested || result.Origin != OriginAgent || result.SourceProfileSnapshotID == nil || *result.SourceProfileSnapshotID != profiles.snapshot || len(result.Tags) != 1 {
		t.Fatalf("profile-backed writing: %+v %v", result, err)
	}
	before := r.creates
	req.DocumentType = TypeRecommendationLetter
	if _, err := s.CreateSample(ctx, req, true); err != ErrSourceRequired || r.creates != before {
		t.Fatal("recommendation without base allowed", err)
	}
	source := uuid.New()
	r.items[source] = &WritingSample{ID: source, UserID: owner, Status: StatusSource}
	req.SourceSampleIDs = []uuid.UUID{source}
	result, err = s.CreateSample(ctx, req, true)
	if err != nil || len(result.SourceSampleIDs) != 1 || result.SourceSampleIDs[0] != source {
		t.Fatal("base provenance lost", err)
	}
	req.DocumentType = TypeEmail
	req.SourceSampleIDs = nil
	if _, err := s.CreateSample(writingContext(owner, principal.ScopeWriting), req, true); err != principal.ErrAgentScopeRequired {
		t.Fatal("missing profiles scope allowed", err)
	}
	profiles.owner = uuid.New()
	if _, err := s.CreateSample(ctx, req, true); err != ErrInvalidProfile {
		t.Fatal("foreign profile allowed", err)
	}
	req.SourceProfileID = nil
	req.SourceSampleIDs = []uuid.UUID{uuid.New()}
	if _, err := s.CreateSample(ctx, req, true); err != ErrInvalidSource {
		t.Fatal("foreign writing allowed", err)
	}
	if _, err := s.GetSample(writingContext(uuid.New(), principal.ScopeWriting), source); err != ErrSampleNotFound {
		t.Fatal("foreign read allowed", err)
	}
	if _, err := s.GetSample(writingContext(owner), source); err != principal.ErrAgentScopeRequired {
		t.Fatal("unscoped read allowed", err)
	}
}
func TestAgentsReviseOnlySuggestedAgentWriting(t *testing.T) {
	owner, id := uuid.New(), uuid.New()
	r := &writingMemory{owner: owner, items: map[uuid.UUID]*WritingSample{}, sources: map[uuid.UUID][]uuid.UUID{}}
	s := NewService(r)
	ctx := writingContext(owner, principal.ScopeWriting)
	content := "Revised supported text"
	approved := StatusApproved
	for _, test := range []struct{ origin, status string }{{OriginUser, StatusSource}, {OriginAgent, StatusApproved}, {OriginAgent, StatusSuggested}} {
		r.items[id] = &WritingSample{ID: id, UserID: owner, Origin: test.origin, Status: test.status, Content: "Before"}
		_, err := s.UpdateSample(ctx, id, UpdateSampleRequest{Content: &content})
		if test.status == StatusSuggested {
			if err != nil || r.items[id].Status != StatusSuggested {
				t.Fatal(err)
			}
		} else if err != ErrDraftReadOnly {
			t.Fatal("source/approved overwrite", err)
		}
	}
	if _, err := s.UpdateSample(ctx, id, UpdateSampleRequest{Status: &approved}); err != ErrDraftReadOnly {
		t.Fatal("agent approved writing", err)
	}
	if _, err := s.AgentListSampleMetadata(ctx, SampleMetadataQuery{Limit: 20}); err != nil || r.metadataOwner != owner {
		t.Fatal("metadata ownership", err)
	}
	if _, err := s.AgentListSampleMetadata(ctx, SampleMetadataQuery{Limit: 51}); err != ErrInvalidQuery {
		t.Fatal(err)
	}
	if _, err := s.MatchSamples(ctx, MatchSamplesRequest{Limit: -1}); err != ErrInvalidQuery {
		t.Fatal("unsafe match limit", err)
	}
}
