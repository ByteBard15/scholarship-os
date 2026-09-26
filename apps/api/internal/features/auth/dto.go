package auth

import (
	"time"

	"github.com/example/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email                string  `json:"email" validate:"required,email"`
	Password             string  `json:"password" validate:"required,min=10"`
	DisplayName          *string `json:"displayName"`
	CreateDefaultProfile *bool   `json:"createDefaultProfile"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=10"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" validate:"required,min=10"`
}

type LoginResponse struct {
	AccessToken string       `json:"accessToken"`
	TokenType   string       `json:"tokenType"`
	ExpiresAt   time.Time    `json:"expiresAt"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"displayName,omitempty"`
	IsActive    bool       `json:"isActive"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func userResponse(value *user.User) UserResponse {
	return UserResponse{ID: value.ID, Email: value.Email, DisplayName: value.DisplayName, IsActive: value.IsActive, LastLoginAt: value.LastLoginAt, CreatedAt: value.CreatedAt}
}

type CreateAgentCredentialRequest struct {
	UserID    *uuid.UUID `json:"userId" validate:"required"`
	Name      string     `json:"name" validate:"required,max=255"`
	Scopes    []string   `json:"scopes" validate:"required,min=1,dive,required"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type AgentCredentialResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *uuid.UUID `json:"userId,omitempty"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"keyPrefix"`
	Scopes     []string   `json:"scopes"`
	IsActive   bool       `json:"isActive"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type CreatedAgentCredentialResponse struct {
	AgentCredentialResponse
	APIKey string `json:"apiKey"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"displayName"`
}
