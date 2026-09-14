package reviewhttp_test

import (
	"context"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestListCourseReviews(t *testing.T) {
	Convey("GET /api/v1/courses/{id}/reviews (public — filter always ignored)", t, func() {
		var svcErr error
		var gotFilter reviewdomain.ReviewFilter
		svc := &mockService{
			getCourseReviews: func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
				gotFilter = f
				return []*reviewdomain.CourseReview{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/courses/"+validCourseID+"/reviews")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (public route)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("course not found → 404", func() {
			svcErr = reviewdomain.ErrCourseNotFound
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with reviews", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["reviews"], ShouldNotBeNil)
		})

		Convey("No query params → zero-value filter reaches service", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotFilter, ShouldResemble, reviewdomain.ReviewFilter{})
		})

		Convey("?rating=3&op=gte is ignored — public route never filters", func() {
			w := testutil.ServeHTTP(mux, newReq("", map[string]string{"rating": "3"}))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotFilter, ShouldResemble, reviewdomain.ReviewFilter{})
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestListContentReviews(t *testing.T) {
	Convey("GET /api/v1/content/{id}/reviews (public — filter always ignored)", t, func() {
		var svcErr error
		var gotFilter reviewdomain.ReviewFilter
		svc := &mockService{
			getContentReviews: func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
				gotFilter = f
				return []*reviewdomain.ContentReview{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/content/"+validContentID+"/reviews")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (public route)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("content item not found → 404", func() {
			svcErr = reviewdomain.ErrContentItemNotFound
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with reviews", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["reviews"], ShouldNotBeNil)
		})

		Convey("?rating=1 is ignored — public route never filters", func() {
			w := testutil.ServeHTTP(mux, newReq("", map[string]string{"rating": "1"}))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotFilter, ShouldResemble, reviewdomain.ReviewFilter{})
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestListCourseReviewsAdmin(t *testing.T) {
	Convey("GET /api/v1/admin/courses/{id}/reviews (rating filter supported)", t, func() {
		var svcErr error
		var gotFilter reviewdomain.ReviewFilter
		svc := &mockService{
			getCourseReviews: func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
				gotFilter = f
				return []*reviewdomain.CourseReview{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/courses/"+validCourseID+"/reviews")
		mux, newReq := f.mux, f.newReq

		// No MustUserFromContext call in this handler — role-gating happens entirely in
		// adminChain middleware (not exercised here, same as other list/stats handlers).

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("No query params → zero-value filter reaches service", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotFilter, ShouldResemble, reviewdomain.ReviewFilter{})
		})

		Convey("?rating=3&op=gte query params parsed into filter", func() {
			w := testutil.ServeHTTP(mux, newReq("", map[string]string{"rating": "3"}))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotFilter.Rating, ShouldEqual, 3)
		})

		Convey("Valid request → 200 with reviews", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["reviews"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestListContentReviewsAdmin(t *testing.T) {
	Convey("GET /api/v1/admin/content/{id}/reviews (rating filter supported)", t, func() {
		var svcErr error
		var gotFilter reviewdomain.ReviewFilter
		svc := &mockService{
			getContentReviews: func(_ context.Context, _ pagination.Params, _ string, f reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
				gotFilter = f
				return []*reviewdomain.ContentReview{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/content/"+validContentID+"/reviews")
		mux, newReq := f.mux, f.newReq

		// No MustUserFromContext call in this handler — role-gating happens entirely in
		// adminChain middleware (not exercised here, same as other list/stats handlers).

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("?op=lte query param parsed into filter", func() {
			w := testutil.ServeHTTP(mux, newReq("", map[string]string{"op": "lte"}))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotFilter.Op, ShouldEqual, "lte")
		})

		Convey("Valid request → 200 with reviews", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["reviews"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}
