package auditdomain

import "errors"

var (
	ErrInvalidItemID   = errors.New("invalid item ID")
	ErrInvalidItemType = errors.New("invalid item type")
)
