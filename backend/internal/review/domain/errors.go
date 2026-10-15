package reviewdomain

import "errors"

// Domain error sentinels for the review module — mapped to HTTP status codes in
// transport/http/errors.go.
var (
	ErrContentItemNotFound = errors.New("contentItem not found")
	ErrCourseNotFound      = errors.New("course not found")
	ErrReviewNotFound      = errors.New("review not found")

	ErrInvalidRating  = errors.New("invalid rating")
	ErrInvalidComment = errors.New("invalid comment")

	ErrInvalidContentItemID = errors.New("invalid contentItem ID")
	ErrInvalidCourseID      = errors.New("invalid course ID")
	ErrInvalidArticleID     = errors.New("invalid article ID")

	ErrInvalidReviewID = errors.New("invalid review ID")

	ErrAlreadyReviewed = errors.New("user already reviewed this item")

	ErrNoPermission = errors.New("user has no permission to perform this action")

	ErrInvalidUserID = errors.New("invalid user ID")
)
