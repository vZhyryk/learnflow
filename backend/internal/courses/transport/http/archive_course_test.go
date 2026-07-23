package coursehttp_test

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArchiveCourse(t *testing.T) {
	Convey("PUT /api/v1/admin/courses/{id}/archive", t, func() {
		var svcErr error

		svc := &mockService{
			archiveCourse: func(_ context.Context, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/courses/11111111-1111-1111-1111-111111111111/archive")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("invalid courseID", func() {
			invalidF := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/courses/---/archive")

			w := testutil.ServeHTTP(invalidF.mux, withUser(invalidF.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
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
			So(body["message"], ShouldEqual, "course was successfully archived")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}
