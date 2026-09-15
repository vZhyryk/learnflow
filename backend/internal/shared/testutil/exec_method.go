package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

// TestExecMethod covers the shared shape of Publish/Archive/Delete-style repository
// methods: Exec, map 0 rows affected to notFoundErr, wrap any other error. Per
// go-testing.md's coverage-per-branch rule and its "consolidate into a shared generic
// helper" guidance once 2+ packages have a near-identical version.
//
// bind receives the MockQueryRunner for the test case and returns the bound repository
// method to exercise.
func TestExecMethod(
	t *testing.T,
	methodName string,
	bind func(*MockQueryRunner) func(ctx context.Context, id, userID string) error,
	notFoundErr error,
) {
	Convey("Given a repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		call := bind(&MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})

		Convey("When it succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(call(context.Background(), "item-123", "user-1"), ShouldBeNil)
		})

		Convey("When no row is matched (not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := call(context.Background(), "unknown", "user-1")
			So(errors.Is(err, notFoundErr), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = ErrDBUnexpected
			err := call(context.Background(), "item-123", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository."+methodName)
		})
	})
}
