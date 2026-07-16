package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"learnflow_backend/internal/infrastructure/db"
)

// NewTestPool opens a pgxpool.Pool against the integration test database (DB_* env
// vars, see docker-compose.tests.yml) and registers pool.Close via t.Cleanup.
func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn, err := db.BuildDSNFromEnv()
	if err != nil {
		t.Fatalf("testutil.NewTestPool: %v", err)
	}

	pool, err := db.InitDatabase(dsn, "1m", "5m", 5, 1)
	if err != nil {
		t.Fatalf("testutil.NewTestPool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// WithTestTx opens a transaction, runs fn, and always rolls back — so integration
// tests never leave rows behind in the shared test database.
func WithTestTx(t *testing.T, pool *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx)) {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("testutil.WithTestTx: begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() //nolint:errcheck // rollback-only helper; error intentionally ignored

	fn(ctx, tx)
}
