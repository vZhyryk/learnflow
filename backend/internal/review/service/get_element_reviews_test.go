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
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			_, err := srv.GetCourseReviews(context.Background(), params, "course-1", noFilter)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.CourseReview{{ID: "review-1"}, {ID: "review-2"}}
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
				return want, nil
			}

			got, err := srv.GetCourseReviews(context.Background(), params, "course-1", noFilter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})

		Convey("success with rating filter — passed through to repository unchanged", func() {
			want := []*reviewdomain.CourseReview{{ID: "review-1", Rating: 5}}
			filter := reviewdomain.ReviewFilter{Op: "gte", Rating: 4}
			var gotFilter reviewdomain.ReviewFilter
			cRepo.getCourseReviewList = func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
				gotFilter = f
				return want, nil
			}

			got, err := srv.GetCourseReviews(context.Background(), params, "course-1", filter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
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
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			_, err := srv.GetContentReviews(context.Background(), params, "content-1", noFilter)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.ContentReview{{ID: "review-1"}}
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
				return want, nil
			}

			got, err := srv.GetContentReviews(context.Background(), params, "content-1", noFilter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})

		Convey("success with rating filter — passed through to repository unchanged", func() {
			want := []*reviewdomain.ContentReview{{ID: "review-1", Rating: 2}}
			filter := reviewdomain.ReviewFilter{Op: "lte", Rating: 2}
			var gotFilter reviewdomain.ReviewFilter
			cRepo.getContentReviewList = func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
				gotFilter = f
				return want, nil
			}

			got, err := srv.GetContentReviews(context.Background(), params, "content-1", filter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
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
			aRepo.getArticleReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			_, err := srv.GetArticleReviews(context.Background(), params, "article-1", noFilter)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			want := []*reviewdomain.ArticleReview{{ID: "review-1"}}
			aRepo.getArticleReviewList = func(_ context.Context, _ pagination.Params, _ string, _ reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error) {
				return want, nil
			}

			got, err := srv.GetArticleReviews(context.Background(), params, "article-1", noFilter)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}
