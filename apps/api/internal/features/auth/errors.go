package auth

import "errors"

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidToken           = errors.New("invalid token")
	ErrTokenExpired           = errors.New("token expired")
	ErrSessionRevoked         = errors.New("session revoked")
	ErrUserInactive           = errors.New("user inactive")
	ErrAgentKeyRevoked        = errors.New("agent key revoked")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrInvalidPassword        = errors.New("password must contain at least 10 characters")
	ErrCredentialNotFound     = errors.New("agent credential not found")
	ErrInvalidAgentScope      = errors.New("invalid agent scope")
	ErrCurrentSessionRequired = errors.New("current user session required")
)
