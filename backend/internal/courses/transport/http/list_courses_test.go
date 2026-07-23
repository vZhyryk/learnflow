package coursehttp_test

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestListCourses(t *testing.T) {
	Convey("GET /api/v1/courses", t, func() {
		var svcErr error

		svc := &mockService{
			getAllCourses: func(_ context.Context, _ coursedomain.CourseStatus, _ pagination.Params) ([]*coursedomain.Course, error) {
				return []*coursedomain.Course{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/courses")
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
			So(body["course_list"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}
