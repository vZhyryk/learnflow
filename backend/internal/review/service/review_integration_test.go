//go:build integration

// Service-level integration tests. Unlike repository integration tests, there is no
// established pattern elsewhere in the codebase for wiring a service against a real
// database — this file establishes it for the review module: the service is built with
// real reviewrepository.Repository + real access.Checker (both backed by the same
// WithTestTx transaction), but with testutil.NoopTransactor instead of the real
// db.PgxTransactor. The real transactor would BEGIN a second, independent transaction
// against the pool — bypassing the outer rollback-only transaction and leaving committed
// rows in the shared test database. NoopTransactor just runs the callback on the same
// ctx/tx, so every write still participates in, and is undone by, the outer rollback.
package reviewservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"learnflow_backend/internal/access"
	reviewdomain "learnflow_backend/internal/review/domain"
	reviewrepository "learnflow_backend/internal/review/repository"
	sharedrepository "learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"

	. "github.com/smartystreets/goconvey/convey"
)

// newIntegrationService wires a Service against real repository + access implementations,
// both backed by tx, with a NoopTransactor (see package doc comment above for why).
func newIntegrationService(tx pgx.Tx) (*Service, *reviewrepository.Repository) {
	repo := &reviewrepository.Repository{BaseRepository: sharedrepository.BaseRepository{DB: tx}}
	return New(repo, repo, repo, testutil.NoopTransactor{}, access.New(tx)), repo
}

func TestCreateCourseReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review service backed by real Postgres and real access checks", t, func() {
		Convey("When the user has active course access and no prior review", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, repo := newIntegrationService(tx)
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})

				err := srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{
					CourseID: courseID, UserID: userID, Rating: 5,
				})

				So(err, ShouldBeNil)
				got, err := repo.GetCourseReviewByUserAndCourseID(ctx, userID, courseID)
				So(err, ShouldBeNil)
				So(got.Rating, ShouldEqual, 5)
			})
		})

		Convey("When the user has no course access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)

				err := srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{
					CourseID: courseID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user's course access has expired", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				created := time.Now().Add(-3 * 24 * time.Hour)
				granted := time.Now().Add(-2 * 24 * time.Hour)
				expires := time.Now().Add(-24 * time.Hour)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{
					AccessType: "purchased", CreatedAt: &created, GrantedAt: &granted, ExpiresAt: &expires,
				})

				err := srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{
					CourseID: courseID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user already reviewed the course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})
				So(srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{CourseID: courseID, UserID: userID, Rating: 3}), ShouldBeNil)

				err := srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{CourseID: courseID, UserID: userID, Rating: 4})

				So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
			})
		})
	})
}

func TestCreateContentReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review service backed by real Postgres and real access checks", t, func() {
		Convey("When the user has direct content access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, repo := newIntegrationService(tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantContentAccess(t, tx, userID, contentID, testutil.AccessGrant{})

				err := srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 4,
				})

				So(err, ShouldBeNil)
				got, err := repo.GetContentReviewByUserAndContentID(ctx, userID, contentID)
				So(err, ShouldBeNil)
				So(got.Rating, ShouldEqual, 4)
			})
		})

		Convey("When the user has access via owning the course the content belongs to (not direct content access)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				courseID := testutil.InsertTestCourse(t, tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				testutil.LinkContentItemToCourse(t, tx, courseID, contentID, 1, true)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})

				err := srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(err, ShouldBeNil)
			})
		})

		Convey("When the user has no access to the content item at all", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)

				err := srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user has active access to a DIFFERENT course that does not contain this content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				unrelatedCourseID := testutil.InsertTestCourse(t, tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, unrelatedCourseID, testutil.AccessGrant{})

				err := srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user's direct content access was revoked", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantContentAccess(t, tx, userID, contentID, testutil.AccessGrant{})
				_, err := tx.Exec(ctx, `UPDATE user_content_access SET status = 'revoked' WHERE user_id = $1 AND content_item_id = $2`, userID, contentID)
				So(err, ShouldBeNil)

				err = srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateContentReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an existing content review", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			srv, repo := newIntegrationService(tx)
			contentID := testutil.InsertTestContentItem(t, tx)
			ownerID := testutil.InsertRandomTestUser(t, tx)
			testutil.GrantContentAccess(t, tx, ownerID, contentID, testutil.AccessGrant{})
			So(srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{ContentID: contentID, UserID: ownerID, Rating: 2}), ShouldBeNil)
			created, err := repo.GetContentReviewByUserAndContentID(ctx, ownerID, contentID)
			So(err, ShouldBeNil)

			Convey("When the owner updates it with active access", func() {
				newRating := 5
				err := srv.UpdateContentReview(ctx, reviewdomain.UpdateContentReviewRequest{
					ReviewID: created.ID, UserID: ownerID, Rating: &newRating,
				})

				So(err, ShouldBeNil)
				got, err := repo.GetContentReviewByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Rating, ShouldEqual, 5)
			})

			Convey("When a different user tries to update it", func() {
				otherUserID := testutil.InsertRandomTestUser(t, tx)
				newRating := 1

				err := srv.UpdateContentReview(ctx, reviewdomain.UpdateContentReviewRequest{
					ReviewID: created.ID, UserID: otherUserID, Rating: &newRating,
				})

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})

			Convey("When the owner's content access was later revoked", func() {
				_, err := tx.Exec(ctx, `UPDATE user_content_access SET status = 'revoked' WHERE user_id = $1 AND content_item_id = $2`, ownerID, contentID)
				So(err, ShouldBeNil)
				newRating := 5

				err = srv.UpdateContentReview(ctx, reviewdomain.UpdateContentReviewRequest{
					ReviewID: created.ID, UserID: ownerID, Rating: &newRating,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteContentReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an existing content review", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			srv, repo := newIntegrationService(tx)
			contentID := testutil.InsertTestContentItem(t, tx)
			ownerID := testutil.InsertRandomTestUser(t, tx)
			testutil.GrantContentAccess(t, tx, ownerID, contentID, testutil.AccessGrant{})
			So(srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{ContentID: contentID, UserID: ownerID, Rating: 2}), ShouldBeNil)
			created, err := repo.GetContentReviewByUserAndContentID(ctx, ownerID, contentID)
			So(err, ShouldBeNil)

			Convey("When the owner deletes it — no content access required for deleting your own review", func() {
				_, err := tx.Exec(ctx, `UPDATE user_content_access SET status = 'revoked' WHERE user_id = $1 AND content_item_id = $2`, ownerID, contentID)
				So(err, ShouldBeNil)

				err = srv.DeleteContentReview(ctx, created.ID, ownerID)

				So(err, ShouldBeNil)
				_, err = repo.GetContentReviewByID(ctx, created.ID)
				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})

			Convey("When a different user tries to delete it", func() {
				otherUserID := testutil.InsertRandomTestUser(t, tx)

				err := srv.DeleteContentReview(ctx, created.ID, otherUserID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateCourseReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an existing course review", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			srv, repo := newIntegrationService(tx)
			courseID := testutil.InsertTestCourse(t, tx)
			ownerID := testutil.InsertRandomTestUser(t, tx)
			testutil.GrantCourseAccess(t, tx, ownerID, courseID, testutil.AccessGrant{})
			So(srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{CourseID: courseID, UserID: ownerID, Rating: 2}), ShouldBeNil)
			created, err := repo.GetCourseReviewByUserAndCourseID(ctx, ownerID, courseID)
			So(err, ShouldBeNil)

			Convey("When the owner updates it with active access", func() {
				newRating := 5
				err := srv.UpdateCourseReview(ctx, reviewdomain.UpdateCourseReviewRequest{
					ReviewID: created.ID, UserID: ownerID, Rating: &newRating,
				})

				So(err, ShouldBeNil)
				got, err := repo.GetCourseReviewByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Rating, ShouldEqual, 5)
			})

			Convey("When a different user tries to update it", func() {
				otherUserID := testutil.InsertRandomTestUser(t, tx)
				newRating := 1

				err := srv.UpdateCourseReview(ctx, reviewdomain.UpdateCourseReviewRequest{
					ReviewID: created.ID, UserID: otherUserID, Rating: &newRating,
				})

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})

			Convey("When the owner's course access was later revoked", func() {
				_, err := tx.Exec(ctx, `UPDATE user_course_access SET status = 'revoked' WHERE user_id = $1 AND course_id = $2`, ownerID, courseID)
				So(err, ShouldBeNil)
				newRating := 5

				err = srv.UpdateCourseReview(ctx, reviewdomain.UpdateCourseReviewRequest{
					ReviewID: created.ID, UserID: ownerID, Rating: &newRating,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteCourseReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an existing course review", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			srv, repo := newIntegrationService(tx)
			courseID := testutil.InsertTestCourse(t, tx)
			ownerID := testutil.InsertRandomTestUser(t, tx)
			testutil.GrantCourseAccess(t, tx, ownerID, courseID, testutil.AccessGrant{})
			So(srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{CourseID: courseID, UserID: ownerID, Rating: 2}), ShouldBeNil)
			created, err := repo.GetCourseReviewByUserAndCourseID(ctx, ownerID, courseID)
			So(err, ShouldBeNil)

			Convey("When the owner deletes it — no course access required for deleting your own review", func() {
				_, err := tx.Exec(ctx, `UPDATE user_course_access SET status = 'revoked' WHERE user_id = $1 AND course_id = $2`, ownerID, courseID)
				So(err, ShouldBeNil)

				err = srv.DeleteCourseReview(ctx, created.ID, ownerID)

				So(err, ShouldBeNil)
				_, err = repo.GetCourseReviewByID(ctx, created.ID)
				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})

			Convey("When a different user tries to delete it", func() {
				otherUserID := testutil.InsertRandomTestUser(t, tx)

				err := srv.DeleteCourseReview(ctx, created.ID, otherUserID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}
