package adminrepository

import (
	"context"
	"encoding/json"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetUsersData(t *testing.T) {
	Convey("Given an admin repository", t, func() {
		var rows *testutil.MockRows
		var queryErr, countErr error
		var gotArgs []any
		total := 42

		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &testutil.MockRow{ScanFn: func(dest ...any) error {
					*testutil.CastInt(dest[0], 0) = total
					return countErr
				}}
			},
			QueryFn: func(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
				gotArgs = args
				return rows, queryErr
			},
		})
		params := pagination.NewParams(2, 10)

		Convey("When the count query fails", func() {
			countErr = testutil.ErrDBUnexpected
			_, _, err := repo.GetUsersData(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetUsersData count")
		})

		Convey("When the list query fails", func() {
			queryErr = testutil.ErrDBUnexpected
			_, _, err := repo.GetUsersData(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetUsersData")
		})

		Convey("When a row fails to scan", func() {
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}}}
			_, _, err := repo.GetUsersData(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}
			_, _, err := repo.GetUsersData(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When the page is empty, it returns a non-nil empty slice that serializes as []", func() {
			rows = &testutil.MockRows{}

			got, gotTotal, err := repo.GetUsersData(context.Background(), params)
			So(err, ShouldBeNil)
			So(gotTotal, ShouldEqual, 42)
			So(got, ShouldNotBeNil)
			So(got, ShouldBeEmpty)
			encoded, marshalErr := json.Marshal(got)
			So(marshalErr, ShouldBeNil)
			So(string(encoded), ShouldEqual, "[]")
		})

		Convey("When rows return 2 users", func() {
			user1, user2 := fakeUserData(1), fakeUserData(2)
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeUserDataScan(user1)},
				{ScanFn: fakeUserDataScan(user2)},
			}}

			got, gotTotal, err := repo.GetUsersData(context.Background(), params)
			So(err, ShouldBeNil)
			So(gotTotal, ShouldEqual, 42)
			So(got, ShouldResemble, []*admindomain.UserData{user1, user2})
			So(gotArgs, ShouldResemble, []any{10, 10})
		})
	})
}

func TestGetUserDataByID(t *testing.T) {
	Convey("Given an admin repository", t, func() {
		var scanFn func(dest ...any) error
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &testutil.MockRow{ScanFn: scanFn}
			},
		})

		Convey("When no user matches", func() {
			scanFn = func(_ ...any) error { return pgx.ErrNoRows }
			got, err := repo.GetUserDataByID(context.Background(), "user-1")
			So(got, ShouldBeNil)
			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database fails", func() {
			scanFn = func(_ ...any) error { return testutil.ErrDBUnexpected }
			_, err := repo.GetUserDataByID(context.Background(), "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetUserDataByID")
		})

		Convey("When the user exists", func() {
			user := fakeUserData(1)
			scanFn = fakeUserDataScan(user)
			got, err := repo.GetUserDataByID(context.Background(), "user-1")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, user)
		})
	})
}

type changeUserMethod struct {
	name string
	call func(*Repository) func(context.Context, string) error
	sql  string
}

var changeUserMethods = []changeUserMethod{
	{"RevokeUserRole", func(r *Repository) func(context.Context, string) error { return r.RevokeUserRole }, revokeUserRoleSQL},
	{"AssignUserRole", func(r *Repository) func(context.Context, string) error { return r.AssignUserRole }, assignUserRoleSQL},
	{"DeleteUser", func(r *Repository) func(context.Context, string) error { return r.DeleteUser }, deleteUserSQL},
	{"RestoreUser", func(r *Repository) func(context.Context, string) error { return r.RestoreUser }, restoreUserSQL},
	{"BlockUser", func(r *Repository) func(context.Context, string) error { return r.BlockUser }, blockUserSQL},
	{"UnBlockUser", func(r *Repository) func(context.Context, string) error { return r.UnBlockUser }, unblockUserSQL},
}

type changeUserFixture struct {
	repo       *Repository
	execTag    pgconn.CommandTag
	execErr    error
	existsErr  error
	userExists bool
	gotQuery   string
	gotArgs    []any
}

func newChangeUserFixture() *changeUserFixture {
	f := &changeUserFixture{}
	f.repo = newTestRepo(&testutil.MockQueryRunner{
		ExecFn: func(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			f.gotQuery, f.gotArgs = sql, args
			return f.execTag, f.execErr
		},
		QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &testutil.MockRow{ScanFn: func(dest ...any) error {
				*testutil.CastBool(dest[0], 0) = f.userExists
				return f.existsErr
			}}
		},
	})

	return f
}

func TestChangeUserMethodsExec(t *testing.T) {
	for _, method := range changeUserMethods {
		Convey("Given "+method.name, t, func() {
			f := newChangeUserFixture()
			call := method.call(f.repo)

			Convey("When a row is updated", func() {
				f.execTag = pgconn.NewCommandTag("UPDATE 1")
				So(call(context.Background(), "user-1"), ShouldBeNil)
				So(f.gotQuery, ShouldEqual, method.sql)
				So(f.gotArgs, ShouldResemble, []any{"user-1"})
			})

			Convey("When the database fails", func() {
				f.execErr = testutil.ErrDBUnexpected
				err := call(context.Background(), "user-1")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "repository."+method.name)
			})
		})
	}
}

func TestChangeUserMethodsNoRowMatched(t *testing.T) {
	for _, method := range changeUserMethods {
		Convey("Given "+method.name+" matching no row", t, func() {
			f := newChangeUserFixture()
			f.execTag = pgconn.NewCommandTag("UPDATE 0")
			call := method.call(f.repo)

			Convey("When the user does not exist", func() {
				err := call(context.Background(), "user-1")
				So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			})

			Convey("When the user is in the wrong state", func() {
				f.userExists = true
				err := call(context.Background(), "user-1")
				So(errors.Is(err, admindomain.ErrInvalidUserState), ShouldBeTrue)
			})

			Convey("When the existence check fails", func() {
				f.existsErr = testutil.ErrDBUnexpected
				err := call(context.Background(), "user-1")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "repository."+method.name+" exists")
			})
		})
	}
}

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
