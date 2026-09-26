package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
)

type memoryProfiles struct {
	items map[uuid.UUID]*ApplicantProfile
}

func newMemoryProfiles() *memoryProfiles {
	return &memoryProfiles{items: map[uuid.UUID]*ApplicantProfile{}}
}
func (m *memoryProfiles) Create(_ context.Context, p *ApplicantProfile) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	m.switchDefault(p)
	copy := *p
	m.items[p.ID] = &copy
	return nil
}
func (m *memoryProfiles) GetByID(_ context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	p, ok := m.items[id]
	if !ok {
		return nil, ErrProfileNotFound
	}
	copy := *p
	return &copy, nil
}
func (m *memoryProfiles) GetFull(c context.Context, id uuid.UUID) (*ApplicantProfile, error) {
	return m.GetByID(c, id)
}
func (m *memoryProfiles) ListByUserID(_ context.Context, userID uuid.UUID) ([]ApplicantProfile, error) {
	out := []ApplicantProfile{}
	for _, p := range m.items {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (m *memoryProfiles) Update(_ context.Context, p *ApplicantProfile) error {
	if _, ok := m.items[p.ID]; !ok {
		return ErrProfileNotFound
	}
	m.switchDefault(p)
	copy := *p
	m.items[p.ID] = &copy
	return nil
}
func (m *memoryProfiles) GetPersonalInfo(context.Context, uuid.UUID) (*ProfilePersonalInfo, error) {
	return nil, ErrPersonalInfoNotFound
}
func (m *memoryProfiles) UpsertPersonalInfo(context.Context, *ProfilePersonalInfo) error { return nil }
func (m *memoryProfiles) switchDefault(p *ApplicantProfile) {
	if !p.IsDefault {
		return
	}
	for _, other := range m.items {
		if other.UserID == p.UserID && other.ID != p.ID {
			other.IsDefault = false
		}
	}
}

type memoryUsers struct{ item *user.User }

func (m memoryUsers) Create(context.Context, *user.User) error { return nil }
func (m memoryUsers) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	if m.item == nil || m.item.ID != id {
		return nil, user.ErrNotFound
	}
	return m.item, nil
}

func TestCreateFirstProfileBecomesDefault(t *testing.T) {
	uid := uuid.New()
	repo := newMemoryProfiles()
	service := NewService(repo, nil, memoryUsers{item: &user.User{ID: uid}})
	p, err := service.Create(context.Background(), uid, CreateProfileRequest{Name: "Primary"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.ID == uuid.Nil {
		t.Fatal("Create() did not assign a UUID")
	}
	if !p.IsDefault {
		t.Fatal("first profile should be default")
	}
}
func TestCreateExplicitDefaultSwitchesExistingDefault(t *testing.T) {
	uid := uuid.New()
	repo := newMemoryProfiles()
	service := NewService(repo, nil, memoryUsers{item: &user.User{ID: uid}})
	first, err := service.Create(context.Background(), uid, CreateProfileRequest{Name: "First"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(context.Background(), uid, CreateProfileRequest{Name: "Second", IsDefault: true})
	if err != nil {
		t.Fatal(err)
	}
	storedFirst, _ := repo.GetByID(context.Background(), first.ID)
	if storedFirst.IsDefault {
		t.Fatal("previous default was not cleared")
	}
	if !second.IsDefault {
		t.Fatal("new profile should be default")
	}
}
func TestUpdateSwitchesDefaultProfile(t *testing.T) {
	uid := uuid.New()
	repo := newMemoryProfiles()
	service := NewService(repo, nil, memoryUsers{item: &user.User{ID: uid}})
	first, _ := service.Create(context.Background(), uid, CreateProfileRequest{Name: "First"})
	second, _ := service.Create(context.Background(), uid, CreateProfileRequest{Name: "Second", IsDefault: true})
	setDefault := true
	updated, err := service.Update(context.Background(), first.ID, UpdateProfileRequest{IsDefault: &setDefault})
	if err != nil {
		t.Fatal(err)
	}
	storedSecond, _ := repo.GetByID(context.Background(), second.ID)
	if !updated.IsDefault || storedSecond.IsDefault {
		t.Fatal("default switch did not leave exactly the requested profile as default")
	}
}
func TestGetProfileNotFound(t *testing.T) {
	service := NewService(newMemoryProfiles(), nil, memoryUsers{})
	_, err := service.Get(context.Background(), uuid.New())
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("Get() error = %v, want ErrProfileNotFound", err)
	}
}
