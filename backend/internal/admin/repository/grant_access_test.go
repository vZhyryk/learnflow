package adminrepository

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

type grantAccessMethod struct {
	name        string
	call        func(r *Repository) func(context.Context, string, string) error
	sql         string
	uniqueIndex string
}

var grantAccessMethods = []grantAccessMethod{
	{
		"GrantUserCourseAccess",
		func(r *Repository) func(context.Context, string, string) error { return r.GrantUserCourseAccess },
		grantUserCourseAccessSQL,
		courseAccessActiveUniqueIndex,
	},
	{
		"GrantUserContentAccess",
		func(r *Repository) func(context.Context, string, string) error { return r.GrantUserContentAccess },
		grantUserContentAccessSQL,
		contentAccessActiveUniqueIndex,
	},
}

func TestGrantAccessMethods(t *testing.T) {
	for _, method := range grantAccessMethods {
		Convey("Given "+method.name, t, func() {
			var gotQuery string
			var gotArgs []any
			var execErr error
			repo := newTestRepo(&testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
					gotQuery, gotArgs = sql, args
					return pgconn.NewCommandTag("INSERT 0 1"), execErr
				},
			})
			call := method.call(repo)

			Convey("When the insert succeeds", func() {
				So(call(context.Background(), "user-1", "item-1"), ShouldBeNil)
				So(gotQuery, ShouldEqual, method.sql)
				So(gotArgs, ShouldResemble, []any{"user-1", "item-1"})
			})

			Convey("When the active-access unique index is violated", func() {
				execErr = &pgconn.PgError{Code: "23505", ConstraintName: method.uniqueIndex}
				err := call(context.Background(), "user-1", "item-1")
				So(errors.Is(err, admindomain.ErrAccessAlreadyGranted), ShouldBeTrue)
			})

			Convey("When another unique constraint is violated", func() {
				execErr = &pgconn.PgError{Code: "23505", ConstraintName: "some_other_unique"}
				err := call(context.Background(), "user-1", "item-1")
				So(errors.Is(err, admindomain.ErrAccessAlreadyGranted), ShouldBeFalse)
				So(err.Error(), ShouldContainSubstring, "repository."+method.name)
			})

			Convey("When the database fails", func() {
				execErr = testutil.ErrDBUnexpected
				err := call(context.Background(), "user-1", "item-1")
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, "repository."+method.name)
			})
		})
	}
}
