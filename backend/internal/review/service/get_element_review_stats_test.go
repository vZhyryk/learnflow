package reviewservice

import (
	"context"
	"errors"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCourseReviewStats(t *testing.T) {
	Convey("Get CourseReviewStats", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(cRepo, nil, nil)

		Convey("repository error", func() {
			cRepo.getCourseReviewStats = func(_ context.Context, _ string) (float64, int, error) {
				return 0, 0, testutil.ErrDBUnexpected
			}

			_, _, err := srv.GetCourseReviewStats(context.Background(), "course-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.getCourseReviewStats = func(_ context.Context, _ string) (float64, int, error) {
				return 4.5, 10, nil
			}

			rating, count, err := srv.GetCourseReviewStats(context.Background(), "course-1")
			So(err, ShouldBeNil)
			So(rating, ShouldEqual, 4.5)
			So(count, ShouldEqual, 10)
		})
	})
}

func TestGetContentReviewStats(t *testing.T) {
	Convey("Get ContentReviewStats", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(nil, cRepo, nil)

		Convey("repository error", func() {
			cRepo.getContentReviewStats = func(_ context.Context, _ string) (float64, int, error) {
				return 0, 0, testutil.ErrDBUnexpected
			}

			_, _, err := srv.GetContentReviewStats(context.Background(), "content-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.getContentReviewStats = func(_ context.Context, _ string) (float64, int, error) {
				return 3.2, 5, nil
			}

			rating, count, err := srv.GetContentReviewStats(context.Background(), "content-1")
			So(err, ShouldBeNil)
			So(rating, ShouldEqual, 3.2)
			So(count, ShouldEqual, 5)
		})
	})
}
