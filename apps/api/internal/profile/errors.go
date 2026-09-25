package profile

import "errors"

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrSectionNotFound      = errors.New("profile section item not found")
	ErrPersonalInfoNotFound = errors.New("personal information not found")
	ErrSkillConflict        = errors.New("skill name already exists for this profile")
)

type validationError struct{ message string }

func (e validationError) Error() string       { return e.message }
func newValidationError(message string) error { return validationError{message: message} }
func isValidationError(err error) bool        { var target validationError; return errors.As(err, &target) }
