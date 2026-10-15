package reviewhttp_test

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCourseReviewStats(t *testing.T) {
	Convey("GET /api/v1/courses/{id}/reviews/stats", t, func() {
		var svcErr error
		var svcRating float64
		var svcCount int
		svc := &mockService{
			getCourseReviewStats: func(_ context.Context, _ string) (float64, int, error) {
				return svcRating, svcCount, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/courses/"+validCourseID+"/reviews/stats")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (public route)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("invalid course id → 422", func() {
			f := newHTTPFixture(svc, http.MethodGet, "/api/v1/courses/---/reviews/stats")
			invalidMux, invalidNewReq := f.mux, f.newReq
			w := testutil.ServeHTTP(invalidMux, invalidNewReq("", nil))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with rating and count", func() {
			svcRating, svcCount = 4.5, 10
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["rating"], ShouldEqual, 4.5)
			So(body["count"], ShouldEqual, float64(10))
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestContentReviewStats(t *testing.T) {
	Convey("GET /api/v1/content/{id}/reviews/stats", t, func() {
		var svcErr error
		var svcRating float64
		var svcCount int
		svc := &mockService{
			getContentReviewStats: func(_ context.Context, _ string) (float64, int, error) {
				return svcRating, svcCount, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/content/"+validContentID+"/reviews/stats")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (public route)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("invalid content id → 422", func() {
			f := newHTTPFixture(svc, http.MethodGet, "/api/v1/content/---/reviews/stats")
			invalidMux, invalidNewReq := f.mux, f.newReq
			w := testutil.ServeHTTP(invalidMux, invalidNewReq("", nil))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with rating and count", func() {
			svcRating, svcCount = 3.2, 7
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["rating"], ShouldEqual, 3.2)
			So(body["count"], ShouldEqual, float64(7))
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestArticleReviewStats(t *testing.T) {
	Convey("GET /api/v1/articles/{id}/reviews/stats", t, func() {
		var svcErr error
		var svcRating float64
		var svcCount int
		svc := &mockService{
			getArticleReviewStats: func(_ context.Context, _ string) (float64, int, error) {
				return svcRating, svcCount, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/articles/"+validArticleID+"/reviews/stats")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (public route)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("invalid article id → 422", func() {
			f := newHTTPFixture(svc, http.MethodGet, "/api/v1/articles/---/reviews/stats")
			invalidMux, invalidNewReq := f.mux, f.newReq
			w := testutil.ServeHTTP(invalidMux, invalidNewReq("", nil))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with rating and count", func() {
			svcRating, svcCount = 4.1, 7
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["rating"], ShouldEqual, 4.1)
			So(body["count"], ShouldEqual, float64(7))
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}
