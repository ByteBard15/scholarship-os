package writing

import "errors"

var (
	ErrSampleNotFound     = errors.New("writing sample not found")
	ErrTagNotFound        = errors.New("writing tag not found")
	ErrUserRequired       = errors.New("user ID is required")
	ErrInvalidApplication = errors.New("application does not belong to the writing sample owner")
	ErrInvalidSource      = errors.New("source writing sample does not belong to the owner")
	ErrTagsRequired       = errors.New("at least one valid tag is required")
)
