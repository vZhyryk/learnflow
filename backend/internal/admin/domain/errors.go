package admindomain

import "errors"

// Announcement validation and lookup errors.
var (
	ErrInvalidID         = errors.New("invalid id")
	ErrInvalidGetType    = errors.New("invalid get type")
	ErrInvalidBody       = errors.New("invalid body")
	ErrInvalidTitle      = errors.New("invalid title")
	ErrInvalidEntityID   = errors.New("invalid entity ID")
	ErrInvalidChannel    = errors.New("invalid channel")
	ErrInvalidEntityType = errors.New("invalid entity type")
	ErrInvalidExpiresAt  = errors.New("invalid expired at")

	ErrEntityDataMisMatch   = errors.New("entity data mismatch")
	ErrAnnouncementNotFound = errors.New("announcement not found")
	ErrAnnouncementApproved = errors.New("announcement already approved")

	ErrUserNotFound        = errors.New("user not found")
	ErrForbiddenUserAction = errors.New("action on this user is not allowed")
	ErrInvalidUserState    = errors.New("user is not in a valid state for this action")
)
