package reviewservice

import (
	"context"
	"errors"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteCourseReview(t *testing.T) {
	Convey("Delete CourseReview", t, func() {
		cRepo := &mockReviewRepo{
			getCourseReviewByID: func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{ID: "review-1", UserID: "user-1"}, nil
			},
			deleteCourseReview: func(_ context.Context, _, _ string) error {
				return nil
			},
		}
		srv := newTestService(cRepo, nil, nil)

		Convey("fetch review error", func() {
			cRepo.getCourseReviewByID = func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.DeleteCourseReview(context.Background(), "review-1", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("not the review owner", func() {
			cRepo.getCourseReviewByID = func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{ID: "review-1", UserID: "someone-else"}, nil
			}

			err := srv.DeleteCourseReview(context.Background(), "review-1", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("repository delete error", func() {
			cRepo.deleteCourseReview = func(_ context.Context, _, _ string) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.DeleteCourseReview(context.Background(), "review-1", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			err := srv.DeleteCourseReview(context.Background(), "review-1", "user-1")
			So(err, ShouldBeNil)
		})
	})
}

func TestDeleteContentReview(t *testing.T) {
	Convey("Delete ContentReview", t, func() {
		cRepo := &mockReviewRepo{
			getContentReviewByID: func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{ID: "review-1", UserID: "user-1"}, nil
			},
			deleteContentReview: func(_ context.Context, _, _ string) error {
				return nil
			},
		}
		srv := newTestService(nil, cRepo, nil)

		Convey("fetch review error", func() {
			cRepo.getContentReviewByID = func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.DeleteContentReview(context.Background(), "review-1", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("not the review owner", func() {
			cRepo.getContentReviewByID = func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{ID: "review-1", UserID: "someone-else"}, nil
			}

			err := srv.DeleteContentReview(context.Background(), "review-1", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("repository delete error", func() {
			cRepo.deleteContentReview = func(_ context.Context, _, _ string) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.DeleteContentReview(context.Background(), "review-1", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			err := srv.DeleteContentReview(context.Background(), "review-1", "user-1")
			So(err, ShouldBeNil)
		})
	})
}

func TestDeleteCourseReviewAdmin(t *testing.T) {
	Convey("Delete CourseReview as admin", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(cRepo, nil, nil)

		Convey("repository delete error", func() {
			cRepo.deleteCourseReview = func(_ context.Context, _, _ string) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.DeleteCourseReviewAdmin(context.Background(), "review-1", "admin-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.deleteCourseReview = func(_ context.Context, _, _ string) error {
				return nil
			}

			err := srv.DeleteCourseReviewAdmin(context.Background(), "review-1", "admin-1")
			So(err, ShouldBeNil)
		})
	})
}

func TestDeleteContentReviewAdmin(t *testing.T) {
	Convey("Delete ContentReview as admin", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(nil, cRepo, nil)

		Convey("repository delete error", func() {
			cRepo.deleteContentReview = func(_ context.Context, _, _ string) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.DeleteContentReviewAdmin(context.Background(), "review-1", "admin-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.deleteContentReview = func(_ context.Context, _, _ string) error {
				return nil
			}

			err := srv.DeleteContentReviewAdmin(context.Background(), "review-1", "admin-1")
			So(err, ShouldBeNil)
		})
	})
}
