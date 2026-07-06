//go:build integration

package authrepository

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"

	. "github.com/smartystreets/goconvey/convey"
)

var uniqueSeq atomic.Int64

// uniqueSuffix returns a monotonically increasing, timestamp-salted suffix
// for building collision-free integration-test fixtures (emails, refresh
// hashes, token hashes) that must be unique within a single test run.
func uniqueSuffix() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), uniqueSeq.Add(1))
}

// fakeUniqueUser returns a fakeUser with a distinct email/ID on every call,
// for integration tests that insert multiple real rows into the same
// unique-active-email index and need them not to collide.
func fakeUniqueUser(now time.Time) *authdomain.User {
	u := fakeUser(now)
	u.ID = fmt.Sprintf("user-%s", uniqueSuffix())
	u.Email = fmt.Sprintf("john+%s@gmail.com", uniqueSuffix())
	return u
}

// newTestUser creates a real user row for integration tests and registers
// cleanup for it and any token rows that might reference it, in FK-safe
// order (tokens before users — three of the four token tables are
// ON DELETE RESTRICT, migration 000001).
func newTestUser(t *testing.T, ctx context.Context, repo *Repository) string {
	t.Helper()

	userId, err := repo.CreateUser(context.Background(), fakeUniqueUser(time.Now()))
	So(err, ShouldBeNil)

	t.Cleanup(func() {
		repo.queryRunner(ctx).Exec(ctx, "DELETE FROM email_verification_tokens WHERE user_id = $1", userId)
		repo.queryRunner(ctx).Exec(ctx, "DELETE FROM password_reset_tokens WHERE user_id = $1", userId)
		repo.queryRunner(ctx).Exec(ctx, "DELETE FROM email_change_tokens WHERE user_id = $1", userId)
		repo.queryRunner(ctx).Exec(ctx, "DELETE FROM account_recovery_tokens WHERE user_id = $1", userId)
		repo.queryRunner(ctx).Exec(ctx, "DELETE FROM users WHERE id = $1", userId)
	})

	return userId
}
