package contentdomain

import "errors"

// Domain error sentinels for the contentItems module — mapped to HTTP status codes in
// transport/http/errors.go.
var (
	ErrContentItemNotFound = errors.New("contentItem not found")

	ErrInvalidSlug               = errors.New("invalid slug")
	ErrInvalidTitle              = errors.New("invalid title")
	ErrInvalidDescription        = errors.New("invalid description")
	ErrInvalidThumbnailURL       = errors.New("invalid Thumbnail URL")
	ErrInvalidEstimatedMinutes   = errors.New("invalid Estimated Minutes")
	ErrInvalidEstimatedPages     = errors.New("invalid Estimated Pages")
	ErrInvalidMedia              = errors.New("invalid media")
	ErrInvalidSeoTitle           = errors.New("invalid seo title")
	ErrInvalidSeoDescription     = errors.New("invalid seo description")
	ErrInvalidOgImageURL         = errors.New("invalid og image url")
	ErrInvalidCanonicalURL       = errors.New("invalid canonical url")
	ErrInvalidAnnouncement       = errors.New("invalid announcement")
	ErrInvalidAnnouncedExpiredAt = errors.New("invalid announced_expired_at")

	ErrInvalidContentItemID = errors.New("invalid contentItem ID")

	ErrInvalidContentItemStatus = errors.New("invalid contentItem status for this operation")

	ErrInvalidGetType = errors.New("invalid get type")
)
