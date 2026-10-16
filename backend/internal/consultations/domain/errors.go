package consultationsdomain

import "errors"

// Errors returned by the consultations module.
var (
	ErrInvalidSlotStatus = errors.New("invalid slot status")
	ErrInvalidSlotID     = errors.New("invalid slot ID")
	ErrDuplicateSlotID   = errors.New("duplicate slot ID")
	ErrEmptySlotUpdate   = errors.New("no slots provided")
	ErrTooManySlots      = errors.New("too many slots in one request")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrDateRangeTooLong  = errors.New("date range is too long")
	ErrInvalidConsultant = errors.New("invalid consultant ID")

	ErrSlotNotFound     = errors.New("slot not found")
	ErrSlotNotAvailable = errors.New("slot is not available")
)
