package testutil

import (
	"context"
	"learnflow_backend/internal/events"

	"github.com/jackc/pgx/v5/pgconn"
)

// NewNoopOutbox returns an OutboxWriter whose Emit always succeeds without touching a real DB.
func NewNoopOutbox() *events.OutboxWriter {
	return events.NewOutboxWriterWithRunner(&MockQueryRunner{
		ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	})
}

// NewFailingOutbox returns an OutboxWriter whose Emit always fails with err.
func NewFailingOutbox(err error) *events.OutboxWriter {
	return events.NewOutboxWriterWithRunner(&MockQueryRunner{
		ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, err
		},
	})
}

// NewCapturingOutbox returns an OutboxWriter that records the Exec args of the last Emit call.
func NewCapturingOutbox(captured *[]any) *events.OutboxWriter {
	return events.NewOutboxWriterWithRunner(&MockQueryRunner{
		ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
			*captured = args
			return pgconn.CommandTag{}, nil
		},
	})
}
