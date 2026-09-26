package profile

import "errors"

var (
	ErrProfileNotFound         = errors.New("profile not found")
	ErrSectionNotFound         = errors.New("profile section item not found")
	ErrPersonalInfoNotFound    = errors.New("personal information not found")
	ErrSkillConflict           = errors.New("skill name already exists for this profile")
	ErrProfileParentRequired   = errors.New("derived profiles require a parent profile")
	ErrInvalidProfileParent    = errors.New("invalid profile parent")
	ErrProfileInheritanceCycle = errors.New("profile inheritance cycle detected")
	ErrProfileOverrideNotFound = errors.New("profile override not found")
	ErrProfileDocumentNotFound = errors.New("profile document not found")
	ErrProfileImportNotFound   = errors.New("profile import not found")
	ErrImportCandidateNotFound = errors.New("import candidate not found")
	ErrImportNotReady          = errors.New("import is not ready to apply")
	ErrInvalidDocumentType     = errors.New("invalid document type")
	ErrInvalidDocument         = errors.New("unsupported or invalid document")
	ErrSnapshotNotFound        = errors.New("profile snapshot not found")
	ErrEvidenceNotFound        = errors.New("profile evidence not found")
)

type validationError struct{ message string }

func (e validationError) Error() string       { return e.message }
func newValidationError(message string) error { return validationError{message: message} }
func isValidationError(err error) bool        { var target validationError; return errors.As(err, &target) }
