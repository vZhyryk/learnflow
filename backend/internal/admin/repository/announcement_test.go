package adminrepository

import (
	"context"
	"errors"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRepository(t *testing.T) {
	Convey("Given a nil connection pool", t, func() {
		Convey("NewRepository returns a non-nil Repository", func() {
			repo := NewRepository(nil)
			So(repo, ShouldNotBeNil)
		})
	})
}

func TestCreateAnnouncement(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given an admin repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeAnnouncement(now)
			row = &testutil.MockRow{ScanFn: fakeAnnouncementScan(expected)}
			got, err := repo.CreateAnnouncement(context.Background(), &admindomain.Announcement{
				Title: expected.Title, Body: expected.Body, CreatedByUserID: expected.CreatedByUserID,
				ExpiresAt: expected.ExpiresAt, EntityID: expected.EntityID, EntityType: expected.EntityType,
				Channels: expected.Channels,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the entity pairing CHECK constraint is violated (pg 23514)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23514", ConstraintName: announcementEntityPairingCheckConstraint}
			}}
			_, err := repo.CreateAnnouncement(context.Background(), &admindomain.Announcement{})
			So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeTrue)
		})

		Convey("When an unrelated check violation occurs (different constraint)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23514", ConstraintName: "some_other_constraint"}
			}}
			_, err := repo.CreateAnnouncement(context.Background(), &admindomain.Announcement{})
			So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When a unique violation occurs (not a CHECK — must not be misclassified)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: announcementEntityPairingCheckConstraint}
			}}
			_, err := repo.CreateAnnouncement(context.Background(), &admindomain.Announcement{})
			So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateAnnouncement(context.Background(), &admindomain.Announcement{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestUpdateAnnouncement(t *testing.T) {
	Convey("Given an admin repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				gotQuery, gotArgs = sql, args
				return execTag, execErr
			},
		})
		announcement := &admindomain.Announcement{ID: "announcement-123"}

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateAnnouncement(context.Background(), announcement), ShouldBeNil)
		})

		Convey("When update succeeds, expires_at is written as the 8th argument", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			expiresAt := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
			announcement.ExpiresAt = expiresAt

			So(repo.UpdateAnnouncement(context.Background(), announcement), ShouldBeNil)

			So(gotQuery, ShouldContainSubstring, "expires_at = $8")
			So(gotArgs, ShouldHaveLength, 8)
			So(gotArgs[7], ShouldResemble, expiresAt)
		})

		Convey("When no row is matched (announcement not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateAnnouncement(context.Background(), announcement)
			So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
		})

		Convey("When the entity pairing CHECK constraint is violated (pg 23514)", func() {
			execErr = &pgconn.PgError{Code: "23514", ConstraintName: announcementEntityPairingCheckConstraint}
			err := repo.UpdateAnnouncement(context.Background(), announcement)
			So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeTrue)
		})

		Convey("When an unrelated check violation occurs (different constraint)", func() {
			execErr = &pgconn.PgError{Code: "23514", ConstraintName: "some_other_constraint"}
			err := repo.UpdateAnnouncement(context.Background(), announcement)
			So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateAnnouncement(context.Background(), announcement)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.UpdateAnnouncement")
		})
	})
}

func TestApproveAnnouncement(t *testing.T) {
	testutil.TestExecMethod(t, "ApproveAnnouncement",
		func(runner *testutil.MockQueryRunner) func(context.Context, string, string) error {
			return newTestRepo(runner).ApproveAnnouncement
		},
		admindomain.ErrAnnouncementNotFound)
}

func TestGetAnnouncementByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given an admin repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the announcement exists", func() {
			expected := fakeAnnouncement(now)
			row = &testutil.MockRow{ScanFn: fakeAnnouncementScan(expected)}
			got, err := repo.GetAnnouncementByID(context.Background(), "announcement-123")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the announcement does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetAnnouncementByID(context.Background(), "unknown")
			So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetAnnouncementByID(context.Background(), "announcement-123")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

// bindAnnouncementList adapts a repository list method into the shape
// testutil.TestListMethod expects: build a repo around the given runner, return the
// bound method.
func bindAnnouncementList(
	call func(*Repository, context.Context, pagination.Params) ([]*admindomain.Announcement, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*admindomain.Announcement, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*admindomain.Announcement, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
			return call(repo, ctx, params)
		}
	}
}

// fakeAnnouncementN builds the nth fake Announcement for TestListMethod's "2 items" case.
func fakeAnnouncementN(n int) *admindomain.Announcement {
	a := fakeAnnouncement(time.Now().UTC().Truncate(time.Second))
	a.ID = fmt.Sprintf("announcement-%d", n)
	return a
}

func TestGetAnnouncements(t *testing.T) {
	testutil.TestListMethod(t, "GetAnnouncements", bindAnnouncementList((*Repository).GetAnnouncements), fakeAnnouncementN, fakeAnnouncementScan)
}

func TestGetUnApprovedAnnouncements(t *testing.T) {
	testutil.TestListMethod(t, "GetUnApprovedAnnouncements", bindAnnouncementList((*Repository).GetUnApprovedAnnouncements), fakeAnnouncementN, fakeAnnouncementScan)
}

func TestGetApprovedAnnouncements(t *testing.T) {
	testutil.TestListMethod(t, "GetApprovedAnnouncements", bindAnnouncementList((*Repository).GetApprovedAnnouncements), fakeAnnouncementN, fakeAnnouncementScan)
}

func TestGetExpiredAnnouncements(t *testing.T) {
	testutil.TestListMethod(t, "GetExpiredAnnouncements", bindAnnouncementList((*Repository).GetExpiredAnnouncements), fakeAnnouncementN, fakeAnnouncementScan)
}

func TestGetPublicAnnouncements(t *testing.T) {
	testutil.TestListMethodWithStringArg(t, "GetPublicAnnouncements",
		func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*admindomain.AnnouncementPublic, error) {
			return newTestRepo(runner).GetPublicAnnouncements
		},
		fakeAnnouncementPublic, fakeAnnouncementPublicScan)
}

func TestGetPublicAnnouncementsArgs(t *testing.T) {
	Convey("Given an admin repository", t, func() {
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
				gotQuery, gotArgs = sql, args
				return &testutil.MockRows{}, nil
			},
		})

		Convey("When listing public announcements for a user", func() {
			params := pagination.NewParams(2, 10)
			_, err := repo.GetPublicAnnouncements(context.Background(), params, "user-1")

			So(err, ShouldBeNil)
			So(gotQuery, ShouldContainSubstring, "'banner' = ANY(a.channels)")
			So(gotArgs, ShouldResemble, []any{"user-1", params.Limit(), params.Offset()})
		})
	})
}
