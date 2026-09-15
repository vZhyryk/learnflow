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

func TestCreateArticleReview(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeArticleReview(now)
			row = &testutil.MockRow{ScanFn: fakeArticleReviewScan(expected)}
			got, err := repo.CreateArticleReview(context.Background(), &reviewdomain.ArticleReview{
				ID: expected.ID, ArticleID: expected.ArticleID, UserID: expected.UserID,
				Rating: expected.Rating, Comment: expected.Comment, CreatedAt: expected.CreatedAt,
				UpdatedAt: expected.UpdatedAt, DeletedAt: expected.DeletedAt,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the user already left the review (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: articleReviewsUserArticleUniqueConstraint}
			}}
			_, err := repo.CreateArticleReview(context.Background(), &reviewdomain.ArticleReview{})
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateArticleReview(context.Background(), &reviewdomain.ArticleReview{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestUpdateArticleReview(t *testing.T) {
	Convey("Given a review repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		item := &reviewdomain.ArticleReview{ID: "article-review-123"}

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateArticleReview(context.Background(), item), ShouldBeNil)
		})

		Convey("When no row is matched (article review not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateArticleReview(context.Background(), item)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateArticleReview(context.Background(), item)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestDeleteArticleReview(t *testing.T) {
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
			So(repo.DeleteArticleReview(context.Background(), "article-review-123", "user-123"), ShouldBeNil)
		})

		Convey("When no row is matched (article review not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.DeleteArticleReview(context.Background(), "article-review-123", "user-123")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.DeleteArticleReview(context.Background(), "article-review-123", "user-123")
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func fakeArticleReviewN(n int) *reviewdomain.ArticleReview {
	item := fakeArticleReview(time.Now().UTC().Truncate(time.Second))
	item.ID = fmt.Sprintf("article-review-%d", n)
	return item
}

func bindArticleReview(
	call func(*Repository, context.Context, pagination.Params, string, reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*reviewdomain.ArticleReview, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params, string) ([]*reviewdomain.ArticleReview, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params, stringArg string) ([]*reviewdomain.ArticleReview, error) {
			return call(repo, ctx, params, stringArg, reviewdomain.ReviewFilter{})
		}
	}
}

func TestGetArticleReviewList(t *testing.T) {
	testutil.TestListMethodWithStringArg(t, "GetArticleReviewList", bindArticleReview((*Repository).GetArticleReviewList), fakeArticleReviewN, fakeArticleReviewScan)
}

func TestGetArticleReviewByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the review exists", func() {
			expected := fakeArticleReview(now)
			row = &testutil.MockRow{ScanFn: fakeArticleReviewScan(expected)}
			got, err := repo.GetArticleReviewByID(context.Background(), "article-review-123")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When no review exists for the given ID", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetArticleReviewByID(context.Background(), "unknown")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetArticleReviewByID(context.Background(), "article-review-123")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetArticleReviewByUserAndArticleID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the user has reviewed the article", func() {
			expected := fakeArticleReview(now)
			row = &testutil.MockRow{ScanFn: fakeArticleReviewScan(expected)}
			got, err := repo.GetArticleReviewByUserAndArticleID(context.Background(), "user_id", "article_id")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the user has not reviewed the article", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetArticleReviewByUserAndArticleID(context.Background(), "user_id", "article_id")
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetArticleReviewByUserAndArticleID(context.Background(), "user_id", "article_id")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetArticleReviewStats(t *testing.T) {
	Convey("Given a review repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When there are reviews", func() {
			row = &testutil.MockRow{ScanFn: func(dest ...any) error {
				*testutil.CastFloat64(dest[0], 0) = 4.5
				*testutil.CastInt(dest[1], 1) = 10
				return nil
			}}
			rating, count, err := repo.GetArticleReviewStats(context.Background(), "article-123")
			So(err, ShouldBeNil)
			So(rating, ShouldEqual, 4.5)
			So(count, ShouldEqual, 10)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, _, err := repo.GetArticleReviewStats(context.Background(), "article-123")
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}
