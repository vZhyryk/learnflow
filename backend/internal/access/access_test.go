package access

import (
	"context"
	"testing"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func rowReturning(exists bool, err error) func(context.Context, string, ...any) pgx.Row {
	return func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &testutil.MockRow{ScanFn: func(dest ...any) error {
			if err != nil {
				return err
			}
			*testutil.CastBool(dest[0], 0) = exists
			return nil
		}}
	}
}

func TestHasAccessCourse(t *testing.T) {
	Convey("Given an access Checker", t, func() {
		Convey("When the query reports access", func() {
			c := New(&testutil.MockQueryRunner{QueryRowFn: rowReturning(true, nil)})

			ok, err := c.HasAccessCourse(context.Background(), "user-1", "course-1")

			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})

		Convey("When the query reports no access", func() {
			c := New(&testutil.MockQueryRunner{QueryRowFn: rowReturning(false, nil)})

			ok, err := c.HasAccessCourse(context.Background(), "user-1", "course-1")

			So(err, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})

		Convey("When the query fails, access is denied and the error is wrapped", func() {
			c := New(&testutil.MockQueryRunner{QueryRowFn: rowReturning(true, testutil.ErrDBUnexpected)})

			ok, err := c.HasAccessCourse(context.Background(), "user-1", "course-1")

			So(ok, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "access.HasAccessCourse")
		})

		Convey("It runs the course query with (userID, courseID)", func() {
			var gotQuery string
			var gotArgs []any
			c := New(&testutil.MockQueryRunner{QueryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
				gotQuery, gotArgs = sql, args
				return rowReturning(true, nil)(ctx, sql, args...)
			}})

			_, err := c.HasAccessCourse(context.Background(), "user-1", "course-1")

			So(err, ShouldBeNil)
			So(gotQuery, ShouldEqual, HasAccessCourseQuery)
			So(gotArgs, ShouldResemble, []any{"user-1", "course-1"})
		})
	})
}

func TestHasAccessContent(t *testing.T) {
	Convey("Given an access Checker", t, func() {
		Convey("When the query reports access", func() {
			c := New(&testutil.MockQueryRunner{QueryRowFn: rowReturning(true, nil)})

			ok, err := c.HasAccessContent(context.Background(), "user-1", "content-1")

			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})

		Convey("When the query fails, access is denied and the error is wrapped", func() {
			c := New(&testutil.MockQueryRunner{QueryRowFn: rowReturning(true, testutil.ErrDBUnexpected)})

			ok, err := c.HasAccessContent(context.Background(), "user-1", "content-1")

			So(ok, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "access.HasAccessContent")
		})

		Convey("It runs the content query with (userID, contentID)", func() {
			var gotQuery string
			var gotArgs []any
			c := New(&testutil.MockQueryRunner{QueryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
				gotQuery, gotArgs = sql, args
				return rowReturning(true, nil)(ctx, sql, args...)
			}})

			_, err := c.HasAccessContent(context.Background(), "user-1", "content-1")

			So(err, ShouldBeNil)
			So(gotQuery, ShouldEqual, HasAccessContentQuery)
			So(gotArgs, ShouldResemble, []any{"user-1", "content-1"})
		})
	})
}
