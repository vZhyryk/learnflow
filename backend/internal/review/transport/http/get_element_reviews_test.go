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
	Convey("GET /api/v1/courses/{id}/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			getCourseReviews: func(_ context.Context, _ pagination.Params, _ string) ([]*reviewdomain.CourseReview, error) {
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

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestListContentReviews(t *testing.T) {
	Convey("GET /api/v1/content/{id}/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			getContentReviews: func(_ context.Context, _ pagination.Params, _ string) ([]*reviewdomain.ContentReview, error) {
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

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}
