package coursehttp_test

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateCourse(t *testing.T) {
	Convey("PUT /api/v1/admin/courses", t, func() {
		var svcErr error
		validBody := `{"id":"11111111-1111-1111-1111-111111111111"}`

		svc := &mockService{
			updateCourse: func(_ context.Context, _ coursedomain.UpdateCourseRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/courses")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("Empty body", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Missing course id fails validation → 400", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service returns an error", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}
