package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"learnflow_backend/internal/infrastructure/db"
)

// NewTestPool opens a pgxpool.Pool against the integration test database (DB_* env
// vars, see docker-compose.tests.yml) and registers pool.Close via t.Cleanup. Fixtures
// built on this pool commit real rows (unlike WithTestTx), so it refuses to run against
// a DB_NAME that doesn't look like a test database.
func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if !strings.Contains(strings.ToLower(os.Getenv("DB_NAME")), "test") {
		t.Fatalf("testutil.NewTestPool: DB_NAME=%q doesn't look like a test database (must contain \"test\")", os.Getenv("DB_NAME"))
	}

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
// Accepts a pgx.Tx (rolled back) or a pool (committed — caller must clean up).
func InsertTestUser(t *testing.T, q db.QueryRunner, email string) string {
	t.Helper()

	var id string
	if err := q.QueryRow(context.Background(), insertTestUserSQL, email, dummyUserPasswordHash).Scan(&id); err != nil {
		t.Fatalf("testutil.InsertTestUser: %v", err)
	}
	return id
}

// InsertRandomTestUser inserts an active user with a unique random email.
func InsertRandomTestUser(t *testing.T, q db.QueryRunner) string {
	t.Helper()

	return InsertTestUser(t, q, RandomTestEmail(t, "test-user"))
}

// RandomTestSlug generates a unique slug, prefixed by the caller's package/purpose
// (e.g. "content-repo-integration"). Mirrors RandomTestEmail's shape for slug columns.
func RandomTestSlug(t *testing.T, prefix string) string {
	t.Helper()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("testutil.RandomTestSlug: %v", err)
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(buf))
}

const insertTestCourseSQL = `
	INSERT INTO courses (slug, title, status, created_by_user_id)
	VALUES ($1, 'Integration Test Course', 'draft', $2)
	RETURNING id`

// InsertTestCourse inserts a minimal draft course (and its owning user) to satisfy a
// courses(id) foreign key.
func InsertTestCourse(t *testing.T, q db.QueryRunner) string {
	t.Helper()

	userID := InsertTestUser(t, q, RandomTestEmail(t, "course-fixture"))
	var id string
	if err := q.QueryRow(context.Background(), insertTestCourseSQL, RandomTestSlug(t, "course-fixture"), userID).Scan(&id); err != nil {
		t.Fatalf("testutil.InsertTestCourse: %v", err)
	}
	return id
}

const insertTestContentItemSQL = `
	INSERT INTO content_items (slug, title, content_type, status, created_by_user_id)
	VALUES ($1, 'Integration Test Content', 'video', 'draft', $2)
	RETURNING id`

// InsertTestContentItem inserts a minimal draft content item (and its owning user) to
// satisfy a content_items(id) foreign key.
func InsertTestContentItem(t *testing.T, q db.QueryRunner) string {
	t.Helper()

	userID := InsertTestUser(t, q, RandomTestEmail(t, "content-fixture"))
	var id string
	if err := q.QueryRow(context.Background(), insertTestContentItemSQL, RandomTestSlug(t, "content-fixture"), userID).Scan(&id); err != nil {
		t.Fatalf("testutil.InsertTestContentItem: %v", err)
	}
	return id
}
