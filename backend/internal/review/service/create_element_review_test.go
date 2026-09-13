package reviewservice

import (
	"context"
	"errors"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCourseReview(t *testing.T) {
	Convey("Create CourseReview", t, func() {
		cRepo := &mockReviewRepo{}
		accessChecker := &mockAccessChecker{
			hasAccessCourse: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(cRepo, nil, accessChecker)

		Convey("No permission", func() {
			accessChecker.hasAccessCourse = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, nil
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
		})

		Convey("access error", func() {
			accessChecker.hasAccessCourse = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetCourseReviewByUserAndCourseID error", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetCourseReviewByUserAndCourseID collision", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("CreateCourseReview error", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateContentReview(t *testing.T) {
	Convey("Create ContentReview", t, func() {
		cRepo := &mockReviewRepo{}
		accessChecker := &mockAccessChecker{
			hasAccessContent: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(nil, cRepo, accessChecker)

		Convey("No permission", func() {
			accessChecker.hasAccessContent = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, nil
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
		})

		Convey("access error", func() {
			accessChecker.hasAccessContent = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetContentReviewByUserAndContentID error", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetContentReviewByUserAndContentID collision", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("CreateContentReview error", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateCourseReviewAdmin(t *testing.T) {
	Convey("Create CourseReview as admin", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(cRepo, nil, nil)

		Convey("GetCourseReviewByUserAndCourseID error", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetCourseReviewByUserAndCourseID collision", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("CreateCourseReview error", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateContentReviewAdmin(t *testing.T) {
	Convey("Create ContentReview as admin", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(nil, cRepo, nil)

		Convey("GetContentReviewByUserAndContentID error", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetContentReviewByUserAndContentID collision", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})

		Convey("CreateContentReview error", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			}

			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}
