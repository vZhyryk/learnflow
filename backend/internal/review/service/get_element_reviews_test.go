package reviewservice

import (
	"context"
	"errors"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCourseReviews(t *testing.T) {
	Convey("Get CourseReviews", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(cRepo, nil, nil)
		params := pagination.NewParams(1, 20)

		Convey("repository error", func() {
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string) ([]*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			_, err := srv.GetCourseReviews(context.Background(), params, "course-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.CourseReview{{ID: "review-1"}, {ID: "review-2"}}
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string) ([]*reviewdomain.CourseReview, error) {
				return want, nil
			}

			got, err := srv.GetCourseReviews(context.Background(), params, "course-1")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}

func TestGetContentReviews(t *testing.T) {
	Convey("Get ContentReviews", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(nil, cRepo, nil)
		params := pagination.NewParams(1, 20)

		Convey("repository error", func() {
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string) ([]*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			_, err := srv.GetContentReviews(context.Background(), params, "content-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.ContentReview{{ID: "review-1"}}
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string) ([]*reviewdomain.ContentReview, error) {
				return want, nil
			}

			got, err := srv.GetContentReviews(context.Background(), params, "content-1")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}
