package authservice

import (
	"context"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"

	"github.com/jackc/pgx/v5/pgconn"
)

// newTestService assembles a Service from mocks. Any nil repo/outbox/redis argument
// is replaced with a zero-value mock (fields left unset will panic if the code under
// test actually calls them, which surfaces missing test setup immediately).
func newTestService(uRepo *mockUserRepo, sRepo *mockSessionRepo, tRepo *mockTokenRepo, outbox *events.OutboxWriter, redisClient *mockRedis) *Service {
	if uRepo == nil {
		uRepo = &mockUserRepo{}
	}
	if sRepo == nil {
		sRepo = &mockSessionRepo{}
	}
	if tRepo == nil {
		tRepo = &mockTokenRepo{}
	}

	srv, err := New(
		Repos{UserRepo: uRepo, SessionRepo: sRepo, TokenRepo: tRepo, Transactor: &mockTransactor{}},
		Utils{
			Token:       tokens.NewTokens("test-secret", "", "learnflow", "learnflow-users"),
			Outbox:      outbox,
			RedisClient: redisClient,
		},
		Options{BcryptCost: 4},
	)
	if err != nil {
		panic(err)
	}
	return srv
}

// newNoopOutbox returns an OutboxWriter whose Emit always succeeds without touching a real DB.
func newNoopOutbox() *events.OutboxWriter {
	return events.NewOutboxWriterWithRunner(&mockQueryRunner{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	})
}

// newFailingOutbox returns an OutboxWriter whose Emit always fails with err.
func newFailingOutbox(err error) *events.OutboxWriter {
	return events.NewOutboxWriterWithRunner(&mockQueryRunner{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, err
		},
	})
}

// newCapturingOutbox returns an OutboxWriter that records the Exec args of the last Emit call.
func newCapturingOutbox(captured *[]any) *events.OutboxWriter {
	return events.NewOutboxWriterWithRunner(&mockQueryRunner{
		execFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
			*captured = args
			return pgconn.CommandTag{}, nil
		},
	})
}
