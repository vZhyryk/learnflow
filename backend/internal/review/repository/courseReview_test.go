package reviewrepository

import (
	"context"
	"errors"
	reviewdomain "learnflow_backend/internal/review/domain"
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

func TestCreateCourseReview(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeCourseReview(now)
			row = &testutil.MockRow{ScanFn: fakeCourseReviewScan(expected)}
			got, err := repo.CreateCourseReview(context.Background(), &reviewdomain.CourseReview{
				ID: expected.ID, CourseID: expected.CourseID, UserID: expected.UserID,
				Rating: expected.Rating, Comment: expected.Comment, CreatedAt: expected.CreatedAt,
				UpdatedAt: expected.UpdatedAt, DeletedAt: expected.DeletedAt,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the user already left the review (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: courseReviewsUserCourseUniqueConstraint}
			}}
			_, err := repo.CreateCourseReview(context.Background(), &reviewdomain.CourseReview{})
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateCourseReview(context.Background(), &reviewdomain.CourseReview{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestUpdateCourseReview(t *testing.T) {
	Convey("Given a review repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		item := &reviewdomain.CourseReview{ID: "content-123"}

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateCourseReview(context.Background(), item), ShouldBeNil)
		})

		Convey("When no row is matched (content item not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateCourseReview(context.Background(), item)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateCourseReview(context.Background(), item)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestDeleteCourseReview(t *testing.T) {
	Convey("Given a review repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		Convey("When delete succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.DeleteCourseReview(context.Background(), "content-123"), ShouldBeNil)
		})

		Convey("When no row is matched (content item not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.DeleteCourseReview(context.Background(), "content-123")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.DeleteCourseReview(context.Background(), "content-123")
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}
