package application

import "errors"

var (
	ErrApplicationNotFound        = errors.New("application not found")
	ErrInvalidApplicationStatus   = errors.New("invalid application status transition")
	ErrApplicationProfileRequired = errors.New("an owned domain profile is required")
	ErrRequirementNotFound        = errors.New("application requirement not found")
	ErrDeadlineNotFound           = errors.New("application deadline not found")
	ErrFundingNotFound            = errors.New("application funding record not found")
	ErrContactNotFound            = errors.New("application contact not found")
	ErrSupervisorNotFound         = errors.New("application supervisor not found")
	ErrURLNotFound                = errors.New("application URL not found")
	ErrTaskNotFound               = errors.New("application task not found")
	ErrTaskParentCycle            = errors.New("task parent cycle detected")
	ErrResearchRunNotFound        = errors.New("research run not found")
	ErrResearchFindingNotFound    = errors.New("research finding not found")
	ErrResearchFindingConflict    = errors.New("research finding conflicts require review")
	ErrResearchNotReady           = errors.New("research findings are not ready to apply")
	ErrInvalidResearchState       = errors.New("invalid research state")
	ErrEvidenceNotFound           = errors.New("requirement evidence not found")
)

type validationError struct{ message string }

func (e validationError) Error() string { return e.message }
func newValidationError(s string) error { return validationError{s} }
