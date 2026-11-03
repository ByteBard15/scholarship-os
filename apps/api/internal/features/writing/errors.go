package writing

import "errors"

var (
	ErrSampleNotFound     = errors.New("writing sample not found")
	ErrTagNotFound        = errors.New("writing tag not found")
	ErrUserRequired       = errors.New("user ID is required")
	ErrInvalidApplication = errors.New("application does not belong to the writing sample owner")
	ErrInvalidSource      = errors.New("source writing sample does not belong to the owner")
	ErrSourceRequired     = errors.New("generated writing needs sourceSampleIds or an owned sourceProfileId; recommendations require sourceSampleIds")
	ErrInvalidProfile     = errors.New("source profile must belong to the writing owner and support snapshots")
	ErrDraftReadOnly      = errors.New("agents may revise only suggested agent-generated writing and cannot approve it")
	ErrInvalidContent     = errors.New("title and content cannot be blank")
	ErrInvalidQuery       = errors.New("limit must be 1 to 50 and offset 0 to 1000000")
	ErrTagsRequired       = errors.New("at least one valid tag is required")
)
