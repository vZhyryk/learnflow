package repository

import (
	"context"
	"errors"
	"testing"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExecUpdateByID(t *testing.T) {
	notFoundErr := errors.New("item not found")

	Convey("Given a BaseRepository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		rep := &BaseRepository{DB: &testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		}}

		Convey("When it succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := ExecUpdateByID(context.Background(), rep, "UPDATE items SET x = $1 WHERE id = $2", "MyMethod", "item-123", notFoundErr)
			So(err, ShouldBeNil)
		})

		Convey("When no row is matched", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := ExecUpdateByID(context.Background(), rep, "UPDATE items SET x = $1 WHERE id = $2", "MyMethod", "unknown", notFoundErr)
			So(errors.Is(err, notFoundErr), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := ExecUpdateByID(context.Background(), rep, "UPDATE items SET x = $1 WHERE id = $2", "MyMethod", "item-123", notFoundErr)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.MyMethod")
		})
	})
}
