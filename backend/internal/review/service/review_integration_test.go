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

	"learnflow_backend/internal/access"
	reviewdomain "learnflow_backend/internal/review/domain"
	reviewrepository "learnflow_backend/internal/review/repository"
	sharedrepository "learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"

	. "github.com/smartystreets/goconvey/convey"
)

func insertTestUser(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "review-service-integration"))
}

func insertTestCourse(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestCourse(t, tx)
}

func insertTestContentItem(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestContentItem(t, tx)
}

const grantCourseAccessSQL = `
	INSERT INTO user_course_access (user_id, course_id, access_type, status, granted_at)
	VALUES ($1, $2, 'admin_granted', 'active', now())`

func grantCourseAccess(t *testing.T, tx pgx.Tx, userID, courseID string) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), grantCourseAccessSQL, userID, courseID); err != nil {
		t.Fatalf("grantCourseAccess: %v", err)
	}
}

const grantContentAccessSQL = `
	INSERT INTO user_content_access (user_id, content_item_id, access_type, status, granted_at)
	VALUES ($1, $2, 'admin_granted', 'active', now())`

func grantContentAccess(t *testing.T, tx pgx.Tx, userID, contentID string) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), grantContentAccessSQL, userID, contentID); err != nil {
		t.Fatalf("grantContentAccess: %v", err)
	}
}

// newIntegrationService wires a Service against real repository + access implementations,
// both backed by tx, with a NoopTransactor (see package doc comment above for why).
func newIntegrationService(tx pgx.Tx) (*Service, *reviewrepository.Repository) {
	repo := &reviewrepository.Repository{BaseRepository: sharedrepository.BaseRepository{DB: tx}}
	return New(repo, repo, testutil.NoopTransactor{}, access.New(tx)), repo
}

func TestCreateCourseReview_ServiceIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a review service backed by real Postgres and real access checks", t, func() {
		Convey("When the user has active course access and no prior review", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, repo := newIntegrationService(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				grantCourseAccess(t, tx, userID, courseID)

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
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)

				err := srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{
					CourseID: courseID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user's course access has expired", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				// created_at is set explicitly in the past too — the granted_at_after_created
				// CHECK requires granted_at >= created_at, and created_at defaults to the real
				// insert-time now() otherwise, which would always be later than a past granted_at.
				_, err := tx.Exec(ctx, `
					INSERT INTO user_course_access (user_id, course_id, access_type, status, created_at, granted_at, expires_at)
					VALUES ($1, $2, 'purchased', 'active', now() - interval '3 days', now() - interval '2 days', now() - interval '1 day')`,
					userID, courseID)
				So(err, ShouldBeNil)

				err = srv.CreateCourseReview(ctx, reviewdomain.CreateCourseReviewRequest{
					CourseID: courseID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user already reviewed the course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				courseID := insertTestCourse(t, tx)
				userID := insertTestUser(t, tx)
				grantCourseAccess(t, tx, userID, courseID)
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
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				grantContentAccess(t, tx, userID, contentID)

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
				courseID := insertTestCourse(t, tx)
				contentID := insertTestContentItem(t, tx)
				_, err := tx.Exec(ctx, `INSERT INTO course_content_items (course_id, content_item_id, position, is_required) VALUES ($1, $2, 1, true)`, courseID, contentID)
				So(err, ShouldBeNil)
				userID := insertTestUser(t, tx)
				grantCourseAccess(t, tx, userID, courseID)

				err = srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(err, ShouldBeNil)
			})
		})

		Convey("When the user has no access to the content item at all", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)

				err := srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user has active access to a DIFFERENT course that does not contain this content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				unrelatedCourseID := insertTestCourse(t, tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				grantCourseAccess(t, tx, userID, unrelatedCourseID)

				err := srv.CreateContentReview(ctx, reviewdomain.CreateContentReviewRequest{
					ContentID: contentID, UserID: userID, Rating: 5,
				})

				So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
			})
		})

		Convey("When the user's direct content access was revoked", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				srv, _ := newIntegrationService(tx)
				contentID := insertTestContentItem(t, tx)
				userID := insertTestUser(t, tx)
				grantContentAccess(t, tx, userID, contentID)
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
			contentID := insertTestContentItem(t, tx)
			ownerID := insertTestUser(t, tx)
			grantContentAccess(t, tx, ownerID, contentID)
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
				otherUserID := insertTestUser(t, tx)
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
			contentID := insertTestContentItem(t, tx)
			ownerID := insertTestUser(t, tx)
			grantContentAccess(t, tx, ownerID, contentID)
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
				otherUserID := insertTestUser(t, tx)

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
			courseID := insertTestCourse(t, tx)
			ownerID := insertTestUser(t, tx)
			grantCourseAccess(t, tx, ownerID, courseID)
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
				otherUserID := insertTestUser(t, tx)
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
			courseID := insertTestCourse(t, tx)
			ownerID := insertTestUser(t, tx)
			grantCourseAccess(t, tx, ownerID, courseID)
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
				otherUserID := insertTestUser(t, tx)

				err := srv.DeleteCourseReview(ctx, created.ID, otherUserID)

				So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
			})
		})
	})
}
