package coursehttp_test

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCourse(t *testing.T) {
	Convey("POST /api/v1/admin/courses", t, func() {
		var svcErr error
		courseID := "course_ID"
		validBody := `{"slug":"test-course","title":"Test Course"}`

		svc := &mockService{
			createCourse: func(_ context.Context, _ coursedomain.CreateCourseRequest) (string, error) {
				return courseID, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/courses")
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

		Convey("Service returns an error", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with course_id", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["course_id"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}
