package user

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *User) error
	GetByID(context.Context, uuid.UUID) (*User, error)
}
