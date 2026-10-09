package notesdomain

import "errors"

// Errors returned by the notes module.
var (
	ErrInvalidTitle         = errors.New("invalid title")
	ErrInvalidDescription   = errors.New("invalid description")
	ErrInvalidBody          = errors.New("invalid body")
	ErrInvalidResourceType  = errors.New("invalid resource type")
	ErrInvalidResourceID    = errors.New("invalid resource ID")
	ErrInvalidUserID        = errors.New("invalid user ID")
	ErrResourceDataMisMatch = errors.New("resource type and resource ID must be provided together")

	ErrNoteNotFound = errors.New("note not found")
)
