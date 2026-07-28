package usersrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetUserProfileByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the profile exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanProfile(now)}
			got, err := repo.GetUserProfileByID(context.Background(), "user-123")
			So(err, ShouldBeNil)
			So(got.UserID, ShouldEqual, "user-123")
			So(*got.FirstName, ShouldEqual, "John")
			So(*got.LastName, ShouldEqual, "Doe")
			So(*got.Country, ShouldEqual, "UA")
			So(got.DateOfBirth, ShouldBeNil)
			So(got.CreatedAt, ShouldEqual, now)
		})

		Convey("When the profile does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserProfileByID(context.Background(), "unknown")
			So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetUserProfileByID(context.Background(), "user-123")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestUpdateUserProfile(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var tag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return tag, execErr
			},
		})

		janeFirstName, janeLastName := "Jane", "Doe"
		profile := &usersdomain.UserProfile{
			UserID:    "user-123",
			FirstName: &janeFirstName,
			LastName:  &janeLastName,
		}

		Convey("When the profile exists and update succeeds", func() {
			tag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateUserProfile(context.Background(), profile), ShouldBeNil)
		})

		Convey("When no row is matched (profile not found)", func() {
			tag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateUserProfile(context.Background(), profile)
			So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBTimeout
			err := repo.UpdateUserProfile(context.Background(), profile)
			testutil.AssertUnexpectedDBError(err, "db timeout")
		})
	})
}
