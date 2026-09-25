package adminrepository

import (
	"context"
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
			So(err.Error(), ShouldContainSubstring, "repository.GetUsersData query")
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
