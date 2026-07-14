package testutil

import (
	"bytes"
	"io"

	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"
)

// NewTestLogger returns a logger that discards all output, for handler tests
// that require a non-nil *logger.Logger but don't assert on log content.
func NewTestLogger() *logger.Logger {
	return logger.New(io.Discard, sanitizer.NewSanitizer("***", 100, nil), logger.LevelFatal)
}

// NewBufferLogger returns a logger writing to a fresh buffer, for tests that need to assert on logged content
func NewBufferLogger(traceLevel logger.Level) (*logger.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return logger.New(&buf, sanitizer.NewSanitizer("***", 2000, nil), traceLevel), &buf
}
