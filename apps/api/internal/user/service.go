package user

import (
	"context"
	"github.com/google/uuid"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Create(ctx context.Context, req CreateRequest) (*User, error) {
	u := &User{Email: strings.ToLower(strings.TrimSpace(req.Email)), DisplayName: req.DisplayName, ExternalAuthID: req.ExternalAuthID}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}
