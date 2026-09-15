package reviewrepository

import (
	"context"
	"errors"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateContentReview(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeContentReview(now)
			row = &testutil.MockRow{ScanFn: fakeContentReviewScan(expected)}
			got, err := repo.CreateContentReview(context.Background(), &reviewdomain.ContentReview{
				ID: expected.ID, ContentID: expected.ContentID, UserID: expected.UserID,
				Rating: expected.Rating, Comment: expected.Comment, CreatedAt: expected.CreatedAt,
				UpdatedAt: expected.UpdatedAt, DeletedAt: expected.DeletedAt,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the user already left the review (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: contentReviewsUserContentUniqueConstraint}
			}}
			_, err := repo.CreateContentReview(context.Background(), &reviewdomain.ContentReview{})
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateContentReview(context.Background(), &reviewdomain.ContentReview{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestUpdateContentReview(t *testing.T) {
	Convey("Given a content repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		item := &reviewdomain.ContentReview{ID: "content-123"}
		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateContentReview(context.Background(), item), ShouldBeNil)
		})

		Convey("When no row is matched (content item not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateContentReview(context.Background(), item)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateContentReview(context.Background(), item)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestDeleteContentReview(t *testing.T) {
	Convey("Given a content repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		Convey("When delete succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.DeleteContentReview(context.Background(), "content-123", "user-123"), ShouldBeNil)
		})

		Convey("When no row is matched (content item not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.DeleteContentReview(context.Background(), "content-123", "user-123")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.DeleteContentReview(context.Background(), "content-123", "user-123")
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestGetContentReviewByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the review is found", func() {
			expected := fakeContentReview(now)
			row = &testutil.MockRow{ScanFn: fakeContentReviewScan(expected)}
			got, err := repo.GetContentReviewByID(context.Background(), expected.ID)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When no review is found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetContentReviewByID(context.Background(), "content-123")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetContentReviewByID(context.Background(), "content-123")
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetContentReviewByUserAndContentID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the review is found", func() {
			expected := fakeContentReview(now)
			row = &testutil.MockRow{ScanFn: fakeContentReviewScan(expected)}
			got, err := repo.GetContentReviewByUserAndContentID(context.Background(), expected.UserID, expected.ContentID)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When no review is found", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetContentReviewByUserAndContentID(context.Background(), "user-123", "content-123")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetContentReviewByUserAndContentID(context.Background(), "user-123", "content-123")
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetContentReviewStats(t *testing.T) {
	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When stats are found", func() {
			row = &testutil.MockRow{ScanFn: func(dest ...any) error {
				*testutil.CastFloat64(dest[0], 0) = 4.5
				*testutil.CastInt(dest[1], 1) = 10
				return nil
			}}
			rating, count, err := repo.GetContentReviewStats(context.Background(), "content-123")
			So(err, ShouldBeNil)
			So(rating, ShouldEqual, 4.5)
			So(count, ShouldEqual, 10)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, _, err := repo.GetContentReviewStats(context.Background(), "content-123")
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func fakeCourse(n int) *reviewdomain.CourseReview {
	item := fakeCourseReview(time.Now().UTC().Truncate(time.Second))
	item.ID = fmt.Sprintf("course-%d", n)
	return item
}

func bindCourseReview(
	call func(*Repository, context.Context, pagination.Params, string, reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*reviewdomain.CourseReview, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*reviewdomain.CourseReview, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params, stringArg string) ([]*reviewdomain.CourseReview, error) {
			return call(repo, ctx, params, stringArg, reviewdomain.ReviewFilter{})
		}
	}
}

func TestGetCourseReviewList(t *testing.T) {
	testutil.TestListMethodWithStringArg(t, "GetCourseReviewList", bindCourseReview((*Repository).GetCourseReviewList), fakeCourse, fakeCourseReviewScan)
}

func fakeContentItem(n int) *reviewdomain.ContentReview {
	item := fakeContentReview(time.Now().UTC().Truncate(time.Second))
	item.ID = fmt.Sprintf("content-%d", n)
	return item
}

func bindContentReview(
	call func(*Repository, context.Context, pagination.Params, string, reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*reviewdomain.ContentReview, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*reviewdomain.ContentReview, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params, stringArg string) ([]*reviewdomain.ContentReview, error) {
			return call(repo, ctx, params, stringArg, reviewdomain.ReviewFilter{})
		}
	}
}

func TestGetContentReviewList(t *testing.T) {
	testutil.TestListMethodWithStringArg(t, "GetContentReviewList", bindContentReview((*Repository).GetContentReviewList), fakeContentItem, fakeContentReviewScan)
}
