package coursedomain

import "errors"

// Domain error sentinels for the courses module — mapped to HTTP status codes in
// transport/http/errors.go.
var (
	ErrCourseNotFound = errors.New("course not found")

	ErrInvalidSlug             = errors.New("invalid slug")
	ErrInvalidTitle            = errors.New("invalid title")
	ErrInvalidDescription      = errors.New("invalid description")
	ErrInvalidThumbnailURL     = errors.New("invalid Thumbnail URL")
	ErrInvalidPreviewVideoURL  = errors.New("invalid Preview Video URL")
	ErrInvalidEstimatedMinutes = errors.New("invalid Estimated Minutes")
	ErrInvalidSeoTitle         = errors.New("invalid seo title")
	ErrInvalidSeoDescription   = errors.New("invalid seo description")
	ErrInvalidOgImageURL       = errors.New("invalid og image url")
	ErrInvalidCanonicalURL     = errors.New("invalid canonical url")

	ErrInvalidCourseID = errors.New("invalid course ID")

	ErrInvalidCourseStatus = errors.New("invalid course status for this operation")

	ErrInvalidGetType = errors.New("invalid get type")
)
