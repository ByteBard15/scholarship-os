package catalog

import "errors"

var (
	ErrInstitutionNotFound = errors.New("institution not found")
	ErrProgrammeNotFound   = errors.New("programme not found")
	ErrScholarshipNotFound = errors.New("scholarship not found")
)
