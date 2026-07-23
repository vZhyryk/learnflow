package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

// dummyUserPasswordHash is a valid bcrypt hash for seeding test users — not a real credential.
const dummyUserPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

const insertTestUserSQL = `
	INSERT INTO users (email, password_hash, role, status)
	VALUES ($1, $2, 'user', 'active')
	RETURNING id`

// RandomTestEmail generates a unique email, prefixed by the caller's package (e.g. "courses-repo").
func RandomTestEmail(t *testing.T, prefix string) string {
	t.Helper()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("testutil.RandomTestEmail: %v", err)
	}
	return fmt.Sprintf("%s-%s@example.com", prefix, hex.EncodeToString(buf))
}

// InsertTestUser inserts a minimal active user row to satisfy a users(id) foreign key.
func InsertTestUser(t *testing.T, tx pgx.Tx, email string) string {
	t.Helper()

	var id string
	if err := tx.QueryRow(context.Background(), insertTestUserSQL, email, dummyUserPasswordHash).Scan(&id); err != nil {
		t.Fatalf("testutil.InsertTestUser: %v", err)
	}
	return id
}
