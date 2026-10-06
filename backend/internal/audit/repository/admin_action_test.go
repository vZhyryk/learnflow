package auditrepository

import (
	"context"
	"errors"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
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
			So(gotArgs, ShouldResemble, []any{"admin-1", auditdomain.ActionBlockUser, auditdomain.TargetUser, "user-1", map[string]any(nil)})
		})

		Convey("When the action carries details", func() {
			action.Details = map[string]any{"course_id": "course-1"}
			So(repo.CreateAdminAction(context.Background(), action), ShouldBeNil)
			So(gotArgs[4], ShouldResemble, map[string]any{"course_id": "course-1"})
		})
	})
}

func TestWasDeletedByAdmin(t *testing.T) {
	Convey("Given an audit repository", t, func() {
		var scanErr error
		var deleted bool
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, args ...any) pgx.Row {
				gotArgs = args
				return &testutil.MockRow{ScanFn: func(dest ...any) error {
					*testutil.CastBool(dest[0], 0) = deleted
					return scanErr
				}}
			},
		})

		Convey("When the query fails, the error is wrapped", func() {
			scanErr = testutil.ErrDBUnexpected
			_, err := repo.WasDeletedByAdmin(context.Background(), "user-1")
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.WasDeletedByAdmin")
		})

		Convey("When no row comes back, it reports false", func() {
			scanErr = pgx.ErrNoRows
			got, err := repo.WasDeletedByAdmin(context.Background(), "user-1")
			So(err, ShouldBeNil)
			So(got, ShouldBeFalse)
		})

		Convey("When an admin deleted the account, it reports true for that target", func() {
			deleted = true
			got, err := repo.WasDeletedByAdmin(context.Background(), "user-1")
			So(err, ShouldBeNil)
			So(got, ShouldBeTrue)
			So(gotArgs, ShouldResemble, []any{"user-1"})
		})
	})
}

func TestNew(t *testing.T) {
	Convey("New returns an Audit backed by the given pool", t, func() {
		So(New(nil), ShouldNotBeNil)
	})
}
