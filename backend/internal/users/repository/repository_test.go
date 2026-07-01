package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

// mockQueryRunner implements db.QueryRunner via function fields.
type mockQueryRunner struct {
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockQueryRunner) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFn(ctx, sql, args...)
}

func (m *mockQueryRunner) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.execFn(ctx, sql, args...)
}

func (m *mockQueryRunner) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	panic("Query not expected in users repository tests")
}

// fakeRow implements pgx.Row for controlled Scan injection.
type fakeRow struct {
	scanFn func(dest ...any) error
}

func (r *fakeRow) Scan(dest ...any) error { return r.scanFn(dest...) }

func newTestRepo(runner *mockQueryRunner) *Repository {
	return &Repository{db: runner}
}

// castStr safely type-asserts a scan destination to *string, panicking with context on failure.
func castStr(v any, idx int) *string {
	s, ok := v.(*string)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *string, got %T", idx, v))
	}
	return s
}

// castPtrStr safely type-asserts a scan destination to **string.
func castPtrStr(v any, idx int) **string {
	s, ok := v.(**string)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected **string, got %T", idx, v))
	}
	return s
}

// castTime safely type-asserts a scan destination to *time.Time.
func castTime(v any, idx int) *time.Time {
	s, ok := v.(*time.Time)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *time.Time, got %T", idx, v))
	}
	return s
}

func fakeProfile(now time.Time) *usersdomain.UserProfile {
	return &usersdomain.UserProfile{
		UserID:      "user-123",
		FirstName:   "John",
		LastName:    "Doe",
		PhoneNumber: "+380991234567",
		Country:     "UA",
		City:        "Kyiv",
		DateOfBirth: nil,
		Gender:      "male",
		UILanguage:  "uk",
		AvatarURL:   "",
		Timezone:    "Europe/Kiev",
		Bio:         "bio text",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func fakeScanProfile(now time.Time) func(dest ...any) error {
	p := fakeProfile(now)
	return func(dest ...any) error {
		*castStr(dest[0], 0) = p.UserID
		*castStr(dest[1], 1) = p.FirstName
		*castStr(dest[2], 2) = p.LastName
		*castStr(dest[3], 3) = p.PhoneNumber
		*castStr(dest[4], 4) = p.Country
		*castStr(dest[5], 5) = p.City
		*castPtrStr(dest[6], 6) = p.DateOfBirth
		*castStr(dest[7], 7) = p.Gender
		*castStr(dest[8], 8) = p.UILanguage
		*castStr(dest[9], 9) = p.AvatarURL
		*castStr(dest[10], 10) = p.Timezone
		*castStr(dest[11], 11) = p.Bio
		*castTime(dest[12], 12) = p.CreatedAt
		*castTime(dest[13], 13) = p.UpdatedAt
		return nil
	}
}

func TestGetUserProfileByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the profile exists", func() {
			row = &fakeRow{scanFn: fakeScanProfile(now)}
			got, err := repo.GetUserProfileByID(context.Background(), "user-123")
			So(err, ShouldBeNil)
			So(got.UserID, ShouldEqual, "user-123")
			So(got.FirstName, ShouldEqual, "John")
			So(got.LastName, ShouldEqual, "Doe")
			So(got.Country, ShouldEqual, "UA")
			So(got.DateOfBirth, ShouldBeNil)
			So(got.CreatedAt, ShouldEqual, now)
		})

		Convey("When the profile does not exist", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserProfileByID(context.Background(), "unknown")
			So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db connection lost") }}
			_, err := repo.GetUserProfileByID(context.Background(), "user-123")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}

func TestUpdateUserProfile(t *testing.T) {
	Convey("Given a users repository", t, func() {
		var tag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return tag, execErr
			},
		})

		profile := &usersdomain.UserProfile{
			UserID:    "user-123",
			FirstName: "Jane",
			LastName:  "Doe",
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
			execErr = errors.New("db timeout")
			err := repo.UpdateUserProfile(context.Background(), profile)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db timeout")
		})
	})
}
