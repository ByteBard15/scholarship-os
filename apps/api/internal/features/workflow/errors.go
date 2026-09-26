package workflow

import "errors"

var (
	ErrResearchTaskNotFound        = errors.New("research task not found")
	ErrResearchTaskLinkNotFound    = errors.New("research task link not found")
	ErrInvalidResearchTaskState    = errors.New("invalid research task state")
	ErrResearchTaskParentCycle     = errors.New("research task parent cycle")
	ErrInvalidResearchTaskParent   = errors.New("invalid research task parent")
	ErrApplicationProposalNotFound = errors.New("application proposal not found")
	ErrInvalidProposalState        = errors.New("invalid application proposal state")
	ErrProposalParentRequired      = errors.New("a master or domain parent profile is required")
	ErrApplicationFieldNotFound    = errors.New("application field not found")
	ErrQuestionnaireNotFound       = errors.New("application questionnaire not found")
	ErrQuestionNotFound            = errors.New("application question not found")
	ErrAnswerNotFound              = errors.New("application answer not found")
	ErrInformationRequestNotFound  = errors.New("information request not found")
	ErrInvalidInformationState     = errors.New("invalid information request state")
	ErrInformationTargetRequired   = errors.New("information request target is required")
	ErrPrefillNotReady             = errors.New("application prefill is not ready")
)

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string   { return e.Message }
func validationError(message string) error { return &ValidationError{Message: message} }
