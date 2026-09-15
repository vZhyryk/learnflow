package reviewhttp_test

import (
	"context"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteCourseReview(t *testing.T) {
	Convey("DELETE /api/v1/courses/reviews/{id}", t, func() {
		var svcErr error
		svc := &mockService{
			deleteCourseReview: func(_ context.Context, _, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/courses/reviews/"+validReviewID)
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("invalid review id → 422", func() {
			f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/courses/reviews/---")
			mux, newReq := f.mux, f.newReq
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("no permission → 403", func() {
			svcErr = reviewdomain.ErrNoPermission
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Course review deleted successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestDeleteContentReview(t *testing.T) {
	Convey("DELETE /api/v1/content/reviews/{id}", t, func() {
		var svcErr error
		svc := &mockService{
			deleteContentReview: func(_ context.Context, _, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/content/reviews/"+validReviewID)
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("invalid review id → 422", func() {
			f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/content/reviews/---")
			mux, newReq := f.mux, f.newReq
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("no permission → 403", func() {
			svcErr = reviewdomain.ErrNoPermission
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Content review deleted successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestDeleteCourseReviewAdmin(t *testing.T) {
	Convey("DELETE /api/v1/admin/courses/reviews/{id}", t, func() {
		var svcErr error
		svc := &mockService{
			deleteCourseReviewAdmin: func(_ context.Context, _, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/admin/courses/reviews/"+validReviewID)
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("invalid review id → 422", func() {
			f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/admin/courses/reviews/---")
			mux, newReq := f.mux, f.newReq
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Course review deleted successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestDeleteContentReviewAdmin(t *testing.T) {
	Convey("DELETE /api/v1/admin/content/reviews/{id}", t, func() {
		var svcErr error
		svc := &mockService{
			deleteContentReviewAdmin: func(_ context.Context, _, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/admin/content/reviews/"+validReviewID)
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("invalid review id → 422", func() {
			f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/admin/content/reviews/---")
			mux, newReq := f.mux, f.newReq
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Content review deleted successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}
