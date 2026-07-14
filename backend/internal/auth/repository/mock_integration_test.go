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

func uniqueSuffix() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), uniqueSeq.Add(1))
}

func fakeUniqueUser(now time.Time) *authdomain.User {
	u := fakeUser(now)
	u.ID = fmt.Sprintf("user-%s", uniqueSuffix())
	u.Email = fmt.Sprintf("john+%s@gmail.com", uniqueSuffix())
	return u
}

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
