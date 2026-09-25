package audit

import (
	"context"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateAdminAction(t *testing.T) {
	Convey("Given an admin repository", t, func() {
		var execErr error
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
				gotArgs = args
				return pgconn.NewCommandTag("INSERT 0 1"), execErr
			},
		})
		action := &auditdomain.AdminAction{
			AdminUserID: "admin-1",
			ActionType:  auditdomain.ActionBlockUser,
			TargetType:  auditdomain.TargetUser,
			TargetID:    "user-1",
		}

		Convey("When the insert fails", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.CreateAdminAction(context.Background(), action)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "audit.CreateAdminAction")
		})

		Convey("When it succeeds", func() {
			So(repo.CreateAdminAction(context.Background(), action), ShouldBeNil)
			So(gotArgs, ShouldResemble, []any{"admin-1", auditdomain.ActionBlockUser, auditdomain.TargetUser, "user-1"})
		})
	})
}
