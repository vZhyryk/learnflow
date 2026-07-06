package authrepository

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateUser(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			row = &testutil.MockRow{ScanFn: func(dest ...any) error {
				*testutil.CastStr(dest[0], 0) = "user-123"
				return nil
			}}
			id, err := repo.CreateUser(context.Background(), &authdomain.User{
				Email: "john@gmail.com", PasswordHash: "hash", Role: authdomain.RoleUser,
			})
			So(err, ShouldBeNil)
			So(id, ShouldEqual, "user-123")
		})

		Convey("When email already exists (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505"}
			}}
			_, err := repo.CreateUser(context.Background(), &authdomain.User{})
			So(errors.Is(err, authdomain.ErrUserAlreadyExists), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateUser(context.Background(), &authdomain.User{})
			assertUnexpectedDBError(err, "db error")
		})

		_ = now
	})
}

func TestCreateUserProfile(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.NewCommandTag("INSERT 1"), fakeErr
			},
		})

		Convey("When creation succeeds", func() {
			firstName, lastName := "John", "Doe"
			err := repo.CreateUserProfile(context.Background(), &authdomain.UserProfile{
				UserID: "user-123", FirstName: &firstName, LastName: &lastName,
			})
			So(err, ShouldBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.CreateUserProfile(context.Background(), &authdomain.UserProfile{})
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetUserByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When user exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUser(now)}
			got, err := repo.GetUserByID(context.Background(), "user-123")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "user-123")
			So(got.Email, ShouldEqual, "john@gmail.com")
			So(got.Role, ShouldEqual, authdomain.UserRole("admin"))
		})

		Convey("When user not found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserByID(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetUserByID(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetUserByEmail(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When user exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUser(now)}
			got, err := repo.GetUserByEmail(context.Background(), "john@gmail.com")
			So(err, ShouldBeNil)
			So(got.Email, ShouldEqual, "john@gmail.com")
		})

		Convey("When user not found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserByEmail(context.Background(), "nobody@example.com")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetUserByEmail(context.Background(), "john@gmail.com")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetDeletedUserByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When deleted user exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUser(now)}
			got, err := repo.GetDeletedUserByID(context.Background(), "user-123")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "user-123")
		})

		Convey("When user not found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetDeletedUserByID(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetDeletedUserByID(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetDeletedUserByEmail(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When deleted user exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUser(now)}
			got, err := repo.GetDeletedUserByEmail(context.Background(), "john@gmail.com")
			So(err, ShouldBeNil)
			So(got.Email, ShouldEqual, "john@gmail.com")
		})

		Convey("When user not found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetDeletedUserByEmail(context.Background(), "nobody@example.com")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetDeletedUserByEmail(context.Background(), "john@gmail.com")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetUserProfileByUserID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When profile exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanProfile(now)}
			got, err := repo.GetUserProfileByUserID(context.Background(), "user-123")
			So(err, ShouldBeNil)
			So(got.UserID, ShouldEqual, "user-123")
			So(*got.FirstName, ShouldEqual, "John")
			So(*got.LastName, ShouldEqual, "Doe")
		})

		Convey("When profile not found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserProfileByUserID(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetUserProfileByUserID(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestRestoreUser(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.RestoreUser(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found (0 rows affected)", func() {
			err := repo.RestoreUser(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When restore succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.RestoreUser(context.Background(), "user-123")
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateStatus(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.UpdateStatus(context.Background(), "user-123", authdomain.StatusBlocked)
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.UpdateStatus(context.Background(), "unknown", authdomain.StatusBlocked)
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateStatus(context.Background(), "user-123", authdomain.StatusActive)
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateRole(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.UpdateRole(context.Background(), "user-123", authdomain.RoleAdmin)
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.UpdateRole(context.Background(), "unknown", authdomain.RoleAdmin)
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateRole(context.Background(), "user-123", authdomain.RoleAdmin)
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateLastLoginAt(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.UpdateLastLoginAt(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.UpdateLastLoginAt(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateLastLoginAt(context.Background(), "user-123")
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdatePasswordHash(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.UpdatePasswordHash(context.Background(), "user-123", "new-hash")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.UpdatePasswordHash(context.Background(), "unknown", "new-hash")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdatePasswordHash(context.Background(), "user-123", "new-hash")
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateEmail(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.UpdateEmail(context.Background(), "user-123", "new@example.com")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.UpdateEmail(context.Background(), "unknown", "new@example.com")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateEmail(context.Background(), "user-123", "new@example.com")
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateEmailVerifiedAt(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.UpdateEmailVerifiedAt(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.UpdateEmailVerifiedAt(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateEmailVerifiedAt(context.Background(), "user-123")
			So(err, ShouldBeNil)
		})
	})
}

func TestDeleteUser(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.DeleteUser(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.DeleteUser(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When delete succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.DeleteUser(context.Background(), "user-123")
			So(err, ShouldBeNil)
		})
	})
}

func TestIncrementFailedLogin(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.IncrementFailedLogin(context.Background(), "user-123", "15 minutes", 5)
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.IncrementFailedLogin(context.Background(), "unknown", "15 minutes", 5)
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When increment succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.IncrementFailedLogin(context.Background(), "user-123", "15 minutes", 5)
			So(err, ShouldBeNil)
		})
	})
}

func TestResetFailedLogin(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := repo.ResetFailedLogin(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When user not found", func() {
			err := repo.ResetFailedLogin(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When reset succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.ResetFailedLogin(context.Background(), "user-123")
			So(err, ShouldBeNil)
		})
	})
}
