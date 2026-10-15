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
		noFilter := reviewdomain.ReviewFilter{}

		Convey("repository error", func() {
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, int, error) {
				return nil, 0, testutil.ErrDBUnexpected
			}

			_, _, err := srv.GetCourseReviews(context.Background(), params, "course-1", noFilter)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.CourseReview{{ID: "review-1"}, {ID: "review-2"}}
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, int, error) {
				return want, 2, nil
			}

			got, total, err := srv.GetCourseReviews(context.Background(), params, "course-1", noFilter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 2)
		})

		Convey("success with rating filter — passed through to repository unchanged", func() {
			want := []*reviewdomain.CourseReview{{ID: "review-1", Rating: 5}}
			filter := reviewdomain.ReviewFilter{Op: "gte", Rating: 4}
			var gotFilter reviewdomain.ReviewFilter
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, int, error) {
				gotFilter = f
				return want, 2, nil
			}

			got, total, err := srv.GetCourseReviews(context.Background(), params, "course-1", filter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 2)
			So(gotFilter, ShouldResemble, filter)
		})
	})
}

func TestGetContentReviews(t *testing.T) {
	Convey("Get ContentReviews", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(nil, cRepo, nil)
		params := pagination.NewParams(1, 20)
		noFilter := reviewdomain.ReviewFilter{}

		Convey("repository error", func() {
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, int, error) {
				return nil, 0, testutil.ErrDBUnexpected
			}

			_, _, err := srv.GetContentReviews(context.Background(), params, "content-1", noFilter)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.ContentReview{{ID: "review-1"}}
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, int, error) {
				return want, 2, nil
			}

			got, total, err := srv.GetContentReviews(context.Background(), params, "content-1", noFilter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 2)
		})

		Convey("success with rating filter — passed through to repository unchanged", func() {
			want := []*reviewdomain.ContentReview{{ID: "review-1", Rating: 2}}
			filter := reviewdomain.ReviewFilter{Op: "lte", Rating: 2}
			var gotFilter reviewdomain.ReviewFilter
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, int, error) {
				gotFilter = f
				return want, 2, nil
			}

			got, total, err := srv.GetContentReviews(context.Background(), params, "content-1", filter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 2)
			So(gotFilter, ShouldResemble, filter)
		})
	})
}

func TestGetArticleReviews(t *testing.T) {
	Convey("Get ArticleReviews", t, func() {
		aRepo := &mockReviewRepo{}
		srv := newTestServiceWithArticleRepo(aRepo)
		params := pagination.NewParams(1, 20)
		noFilter := reviewdomain.ReviewFilter{}

		Convey("repository error", func() {
			aRepo.getArticleReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, int, error) {
				return nil, 0, testutil.ErrDBUnexpected
			}

			_, _, err := srv.GetArticleReviews(context.Background(), params, "article-1", noFilter)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.ArticleReview{{ID: "review-1"}}
			aRepo.getArticleReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, int, error) {
				return want, 2, nil
			}

			got, total, err := srv.GetArticleReviews(context.Background(), params, "article-1", noFilter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 2)
		})
	})
}
