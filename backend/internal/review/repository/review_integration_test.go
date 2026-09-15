//go:build integration

package reviewrepository

import (
	"context"
	"errors"
	"testing"

	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

func insertTestUser(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "review-repo-integration"))
}

func insertTestCourse(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestCourse(t, tx)
}

func insertTestContentItem(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestContentItem(t, tx)
}

func newTestRepository(tx pgx.Tx) *Repository {
	return &Repository{repository.BaseRepository{DB: tx}}
}

// === CreateCourseReview / CreateContentReview ===

func TestCreateCourseReview_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When creating a course review with valid course and user", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				comment := "great course"

				got, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{
					CourseID: courseID,
					UserID:   userID,
					Rating:   5,
					Comment:  &comment,
				})

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.CourseID, ShouldEqual, courseID)
				So(got.UserID, ShouldEqual, userID)
				So(got.Rating, ShouldEqual, 5)
				So(*got.Comment, ShouldEqual, comment)
				So(got.DeletedAt, ShouldBeNil)
			})
		})

		Convey("When the user already has an active review for the course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)

				_, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 4})
				So(err, ShouldBeNil)

				_, err = repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 2})

				So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
			})
		})

		Convey("When course_id does not reference an existing course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				userID := insertTestUser(t, tx)

				_, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{
					CourseID: "00000000-0000-0000-0000-000000000000",
					UserID:   userID,
					Rating:   5,
				})

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503") // foreign_key_violation
			})
		})

		Convey("When rating is outside the 1-5 range (DB-level CHECK, defense-in-depth below domain validation)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)

				_, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 6})

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514") // check_violation
				So(pgErr.ConstraintName, ShouldEqual, "course_reviews_rating_check")
			})
		})

		Convey("When the same user reviews again after their first review was soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)

				first, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)
				So(repo.DeleteCourseReview(ctx, first.ID, userID), ShouldBeNil)

				second, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 5})

				So(err, ShouldBeNil)
				So(second.ID, ShouldNotEqual, first.ID)
			})
		})
	})
}

func TestCreateContentReview_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When creating a content review with valid content item and user", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				comment := "very helpful"

				got, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{
					ContentID: contentID,
					UserID:    userID,
					Rating:    4,
					Comment:   &comment,
				})

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.ContentID, ShouldEqual, contentID)
				So(got.UserID, ShouldEqual, userID)
				So(got.Rating, ShouldEqual, 4)
			})
		})

		Convey("When the user already has an active review for the content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)

				_, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: userID, Rating: 4})
				So(err, ShouldBeNil)

				_, err = repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: userID, Rating: 1})

				So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
			})
		})

		Convey("When content_item_id does not reference an existing content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				userID := insertTestUser(t, tx)

				_, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{
					ContentID: "00000000-0000-0000-0000-000000000000",
					UserID:    userID,
					Rating:    5,
				})

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503")
			})
		})
	})
}

// === UpdateCourseReview / UpdateContentReview ===

func TestUpdateCourseReview_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When updating rating and comment of an existing review", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 2})
				So(err, ShouldBeNil)

				newComment := "updated my mind"
				err = repo.UpdateCourseReview(ctx, &reviewdomain.CourseReview{ID: created.ID, Rating: 5, Comment: &newComment})
				So(err, ShouldBeNil)

				got, err := repo.GetCourseReviewByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Rating, ShouldEqual, 5)
				So(*got.Comment, ShouldEqual, newComment)
			})
		})

		Convey("When the review does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)

				err := repo.UpdateCourseReview(ctx, &reviewdomain.CourseReview{ID: "00000000-0000-0000-0000-000000000000", Rating: 3})

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})

		Convey("When the review is soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 2})
				So(err, ShouldBeNil)
				So(repo.DeleteCourseReview(ctx, created.ID, userID), ShouldBeNil)

				err = repo.UpdateCourseReview(ctx, &reviewdomain.CourseReview{ID: created.ID, Rating: 5})

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateContentReview_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When updating rating and comment of an existing review", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: userID, Rating: 2})
				So(err, ShouldBeNil)

				newComment := "actually pretty good"
				err = repo.UpdateContentReview(ctx, &reviewdomain.ContentReview{ID: created.ID, Rating: 4, Comment: &newComment})
				So(err, ShouldBeNil)

				got, err := repo.GetContentReviewByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Rating, ShouldEqual, 4)
				So(*got.Comment, ShouldEqual, newComment)
			})
		})

		Convey("When the review does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)

				err := repo.UpdateContentReview(ctx, &reviewdomain.ContentReview{ID: "00000000-0000-0000-0000-000000000000", Rating: 3})

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

// === DeleteCourseReview / DeleteContentReview ===

func TestDeleteCourseReview_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When soft-deleting an existing review", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)

				So(repo.DeleteCourseReview(ctx, created.ID, userID), ShouldBeNil)

				_, err = repo.GetCourseReviewByID(ctx, created.ID)
				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})

		Convey("When the review is already deleted (second delete affects 0 rows)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)
				So(repo.DeleteCourseReview(ctx, created.ID, userID), ShouldBeNil)

				err = repo.DeleteCourseReview(ctx, created.ID, userID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})

		Convey("When the review does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)

				err := repo.DeleteCourseReview(ctx, "00000000-0000-0000-0000-000000000000", insertTestUser(t, tx))

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteContentReview_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When soft-deleting an existing review", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)

				So(repo.DeleteContentReview(ctx, created.ID, userID), ShouldBeNil)

				_, err = repo.GetContentReviewByID(ctx, created.ID)
				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})

		Convey("When the review does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)

				err := repo.DeleteContentReview(ctx, "00000000-0000-0000-0000-000000000000", insertTestUser(t, tx))

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

// === GetCourseReviewByID / GetContentReviewByID ===

func TestGetCourseReviewByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When the review exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 4})
				So(err, ShouldBeNil)

				got, err := repo.GetCourseReviewByID(ctx, created.ID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When no review exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)

				_, err := repo.GetCourseReviewByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetContentReviewByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When the review exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: userID, Rating: 4})
				So(err, ShouldBeNil)

				got, err := repo.GetContentReviewByID(ctx, created.ID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When no review exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)

				_, err := repo.GetContentReviewByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

// === GetCourseReviewByUserAndCourseID / GetContentReviewByUserAndContentID ===

func TestGetCourseReviewByUserAndCourseID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When the user has reviewed the course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)

				got, err := repo.GetCourseReviewByUserAndCourseID(ctx, userID, courseID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When the user has not reviewed the course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)

				_, err := repo.GetCourseReviewByUserAndCourseID(ctx, userID, courseID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})

		Convey("When the user's review was soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)
				So(repo.DeleteCourseReview(ctx, created.ID, userID), ShouldBeNil)

				_, err = repo.GetCourseReviewByUserAndCourseID(ctx, userID, courseID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetContentReviewByUserAndContentID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When the user has reviewed the content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				created, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: userID, Rating: 3})
				So(err, ShouldBeNil)

				got, err := repo.GetContentReviewByUserAndContentID(ctx, userID, contentID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When the user has not reviewed the content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)

				_, err := repo.GetContentReviewByUserAndContentID(ctx, userID, contentID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

// === GetCourseReviewList / GetContentReviewList ===

func TestGetCourseReviewList_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a course with multiple reviews, one soft-deleted", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := newTestRepository(tx)
			courseID := insertTestCourse(t, tx)

			active1, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: insertTestUser(t, tx), Rating: 5})
			So(err, ShouldBeNil)
			active2, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: insertTestUser(t, tx), Rating: 3})
			So(err, ShouldBeNil)
			deleted, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: insertTestUser(t, tx), Rating: 1})
			So(err, ShouldBeNil)
			So(repo.DeleteCourseReview(ctx, deleted.ID, insertTestUser(t, tx)), ShouldBeNil)

			Convey("Listing returns only active reviews for that course", func() {
				got, err := repo.GetCourseReviewList(ctx, pagination.NewParams(1, 100), courseID, reviewdomain.ReviewFilter{})

				So(err, ShouldBeNil)
				ids := courseReviewIDs(got)
				So(ids, ShouldContain, active1.ID)
				So(ids, ShouldContain, active2.ID)
				So(ids, ShouldNotContain, deleted.ID)
			})

			Convey("Pagination limits the returned page size", func() {
				got, err := repo.GetCourseReviewList(ctx, pagination.NewParams(1, 1), courseID, reviewdomain.ReviewFilter{})

				So(err, ShouldBeNil)
				So(len(got), ShouldEqual, 1)
			})

			Convey("A different course has no reviews", func() {
				otherCourseID := insertTestCourse(t, tx)

				got, err := repo.GetCourseReviewList(ctx, pagination.NewParams(1, 100), otherCourseID, reviewdomain.ReviewFilter{})

				So(err, ShouldBeNil)
				So(got, ShouldBeEmpty)
			})

			Convey("rating filter (gte) returns only reviews matching the comparison", func() {
				got, err := repo.GetCourseReviewList(ctx, pagination.NewParams(1, 100), courseID, reviewdomain.ReviewFilter{Op: "gte", Rating: 4})

				So(err, ShouldBeNil)
				ids := courseReviewIDs(got)
				So(ids, ShouldContain, active1.ID) // rating 5
				So(ids, ShouldNotContain, active2.ID) // rating 3
				So(ids, ShouldNotContain, deleted.ID)
			})

			Convey("rating filter (eq) matches exactly", func() {
				got, err := repo.GetCourseReviewList(ctx, pagination.NewParams(1, 100), courseID, reviewdomain.ReviewFilter{Op: "eq", Rating: 3})

				So(err, ShouldBeNil)
				ids := courseReviewIDs(got)
				So(ids, ShouldContain, active2.ID)
				So(ids, ShouldNotContain, active1.ID)
			})
		})
	})
}

func TestGetContentReviewList_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content item with multiple reviews, one soft-deleted", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := newTestRepository(tx)
			contentID := insertTestContentItem(t, tx)

			active, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: insertTestUser(t, tx), Rating: 5})
			So(err, ShouldBeNil)
			deleted, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: insertTestUser(t, tx), Rating: 1})
			So(err, ShouldBeNil)
			So(repo.DeleteContentReview(ctx, deleted.ID, insertTestUser(t, tx)), ShouldBeNil)

			Convey("Listing returns only active reviews for that content item", func() {
				got, err := repo.GetContentReviewList(ctx, pagination.NewParams(1, 100), contentID, reviewdomain.ReviewFilter{})

				So(err, ShouldBeNil)
				ids := contentReviewIDs(got)
				So(ids, ShouldContain, active.ID)
				So(ids, ShouldNotContain, deleted.ID)
			})

			Convey("rating filter (lte) returns only reviews matching the comparison", func() {
				got, err := repo.GetContentReviewList(ctx, pagination.NewParams(1, 100), contentID, reviewdomain.ReviewFilter{Op: "lte", Rating: 3})

				So(err, ShouldBeNil)
				ids := contentReviewIDs(got)
				So(ids, ShouldNotContain, active.ID) // rating 5
				So(ids, ShouldNotContain, deleted.ID)
			})
		})
	})
}

// === GetCourseReviewStats / GetContentReviewStats ===

func TestGetCourseReviewStats_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When a course has no reviews", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)

				rating, count, err := repo.GetCourseReviewStats(ctx, courseID)

				So(err, ShouldBeNil)
				So(rating, ShouldEqual, 0)
				So(count, ShouldEqual, 0)
			})
		})

		Convey("When a course has multiple active reviews and one soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				courseID := insertTestCourse(t, tx)

				_, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: insertTestUser(t, tx), Rating: 4})
				So(err, ShouldBeNil)
				_, err = repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: insertTestUser(t, tx), Rating: 2})
				So(err, ShouldBeNil)
				excluded, err := repo.CreateCourseReview(ctx, &reviewdomain.CourseReview{CourseID: courseID, UserID: insertTestUser(t, tx), Rating: 1})
				So(err, ShouldBeNil)
				So(repo.DeleteCourseReview(ctx, excluded.ID, insertTestUser(t, tx)), ShouldBeNil)

				rating, count, err := repo.GetCourseReviewStats(ctx, courseID)

				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
				So(rating, ShouldEqual, 3) // (4+2)/2, excludes the soft-deleted rating of 1
			})
		})
	})
}

func TestGetContentReviewStats_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review repository backed by real Postgres", t, func() {
		Convey("When a content item has no reviews", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)

				rating, count, err := repo.GetContentReviewStats(ctx, contentID)

				So(err, ShouldBeNil)
				So(rating, ShouldEqual, 0)
				So(count, ShouldEqual, 0)
			})
		})

		Convey("When a content item has multiple active reviews", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newTestRepository(tx)
				contentID := insertTestContentItem(t, tx)

				_, err := repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: insertTestUser(t, tx), Rating: 5})
				So(err, ShouldBeNil)
				_, err = repo.CreateContentReview(ctx, &reviewdomain.ContentReview{ContentID: contentID, UserID: insertTestUser(t, tx), Rating: 3})
				So(err, ShouldBeNil)

				rating, count, err := repo.GetContentReviewStats(ctx, contentID)

				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
				So(rating, ShouldEqual, 4)
			})
		})
	})
}

func courseReviewIDs(reviews []*reviewdomain.CourseReview) []string {
	ids := make([]string, 0, len(reviews))
	for _, r := range reviews {
		ids = append(ids, r.ID)
	}
	return ids
}

func contentReviewIDs(reviews []*reviewdomain.ContentReview) []string {
	ids := make([]string, 0, len(reviews))
	for _, r := range reviews {
		ids = append(ids, r.ID)
	}
	return ids
}
