// Package testutil provides shared test doubles for service/repository unit tests.
// It is imported only from _test.go files and never ships in a production binary.
package testutil

import (
	"context"
	"errors"
)

// ErrDBUnexpected is a stand-in for an unexpected persistence-layer failure.
// Use it wherever the test asserts on a wrapped context message (or nothing at
// all), not on this error's own text.
var ErrDBUnexpected = errors.New("db connection lost")

// ErrDB is a stand-in for an unexpected persistence-layer failure whose literal
// text is itself asserted on (repository-layer tests that check
// err.Error() ShouldContainSubstring "db error").
var ErrDB = errors.New("db error")

// ErrDBTimeout is a stand-in for an unexpected persistence-layer failure whose
// literal text is itself asserted on (tests that check
// err.Error() ShouldContainSubstring "db timeout").
var ErrDBTimeout = errors.New("db timeout")

// ErrRedisUnavailable is a stand-in for an unexpected Redis-layer failure
// (SetNX, LPush, Publish, etc.) — the Redis equivalent of ErrDBUnexpected.
var ErrRedisUnavailable = errors.New("redis unavailable")

// AlwaysNil satisfies any mock field shaped func(context.Context, string) error
// that should succeed unconditionally.
func AlwaysNil(_ context.Context, _ string) error {
	return nil
}

// AlwaysFailsDB satisfies any mock field shaped func(context.Context, string) error
// that should fail with ErrDBUnexpected.
func AlwaysFailsDB(_ context.Context, _ string) error {
	return ErrDBUnexpected
}

// AlwaysNil2 satisfies any mock field shaped func(context.Context, string, string) error
// that should succeed unconditionally.
func AlwaysNil2(_ context.Context, _, _ string) error {
	return nil
}

// AlwaysFailsDB2 satisfies any mock field shaped func(context.Context, string, string) error
// that should fail with ErrDBUnexpected.
func AlwaysFailsDB2(_ context.Context, _, _ string) error {
	return ErrDBUnexpected
}
