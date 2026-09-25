package user

import (
	"github.com/google/uuid"
	"time"
)

type CreateRequest struct {
	Email          string  `json:"email" validate:"required,email"`
	DisplayName    *string `json:"displayName"`
	ExternalAuthID *string `json:"externalAuthId"`
}
type Response struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	DisplayName    *string   `json:"displayName,omitempty"`
	ExternalAuthID *string   `json:"externalAuthId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func ToResponse(u *User) Response {
	return Response{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, ExternalAuthID: u.ExternalAuthID, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}
