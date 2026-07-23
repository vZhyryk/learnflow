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
	"github.com/jackc/pgx/v5/pgtype"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{db: runner}
}

// fakeProfile/fakeScanProfile mirror internal/auth/repository/mock_test.go's fixture
// of the same name. Not shared via testutil on purpose: usersdomain.UserProfile and
// authdomain.UserProfile are separate types by design (auth and users are independent
// bounded contexts per Clean Architecture layering), so a generic helper here would
// either need generics (forbidden in domain-adjacent code) or reflection.
func fakeProfile(now time.Time) *usersdomain.UserProfile {
	firstName, lastName, phoneNumber := "John", "Doe", "+380991234567"
	country, city, gender := "UA", "Kyiv", "male"
	timezone, bio := "Europe/Kiev", "bio text"
	avatarURL := ""
	return &usersdomain.UserProfile{
		UserID:      "user-123",
		FirstName:   &firstName,
		LastName:    &lastName,
		PhoneNumber: &phoneNumber,
		Country:     &country,
		City:        &city,
		DateOfBirth: nil,
		Gender:      &gender,
		UILanguage:  "uk",
		AvatarURL:   &avatarURL,
		Timezone:    &timezone,
		Bio:         &bio,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func fakeScanProfile(now time.Time) func(dest ...any) error {
	p := fakeProfile(now)
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = p.UserID
		*testutil.CastPtrStr(dest[1], 1) = p.FirstName
		*testutil.CastPtrStr(dest[2], 2) = p.LastName
		*testutil.CastPtrStr(dest[3], 3) = p.PhoneNumber
		*testutil.CastPtrStr(dest[4], 4) = p.Country
		*testutil.CastPtrStr(dest[5], 5) = p.City
		*testutil.CastPgtypeDate(dest[6], 6) = pgtype.Date{}
		*testutil.CastPtrStr(dest[7], 7) = p.Gender
		*testutil.CastStr(dest[8], 8) = p.UILanguage
		*testutil.CastPtrStr(dest[9], 9) = p.AvatarURL
		*testutil.CastPtrStr(dest[10], 10) = p.Timezone
		*testutil.CastPtrStr(dest[11], 11) = p.Bio
		*testutil.CastTime(dest[12], 12) = p.CreatedAt
		*testutil.CastTime(dest[13], 13) = p.UpdatedAt
		return nil
	}
}

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
