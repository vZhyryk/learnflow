// Package apperrors holds cross-module sentinel errors.
package apperrors

import "errors"

// ErrNotFound reports that a requested entity does not exist.
var ErrNotFound = errors.New("not found")
