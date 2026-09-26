package user

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailConflict = errors.New("email already exists")
)
