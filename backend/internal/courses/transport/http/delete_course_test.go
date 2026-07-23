package coursehttp_test

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteCourse(t *testing.T) {
	Convey("DELETE /api/v1/admin/courses/{id}", t, func() {
		var svcErr error

		svc := &mockService{
			deleteCourse: func(_ context.Context, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/admin/courses/11111111-1111-1111-1111-111111111111")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "course was successfully deleted")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})

		Convey("invalid courseID", func() {
			f := newHTTPFixture(svc, http.MethodDelete, "/api/v1/admin/courses/---")
			mux, newReq := f.mux, f.newReq

			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})
	})
}
