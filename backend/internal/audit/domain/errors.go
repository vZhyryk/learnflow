package auditdomain

import "errors"

// Errors returned when validating audit queries.
var (
	ErrInvalidItemID      = errors.New("invalid item ID")
	ErrInvalidItemType    = errors.New("invalid item type")
	ErrInvalidAdminUserID = errors.New("invalid admin user ID")
	ErrInvalidActionType  = errors.New("invalid action type")
	ErrInvalidDate        = errors.New("invalid date: use RFC 3339, e.g. 2026-10-01T00:00:00Z")
	ErrInvalidDateRange   = errors.New("invalid date range: from must be before to")
)
