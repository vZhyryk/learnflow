package testutil

import (
	"io"

	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"
)

// NewTestLogger returns a logger that discards all output, for handler tests
// that require a non-nil *logger.Logger but don't assert on log content.
func NewTestLogger() *logger.Logger {
	return logger.New(io.Discard, sanitizer.NewSanitizer("***", 100, nil), logger.LevelFatal)
}
