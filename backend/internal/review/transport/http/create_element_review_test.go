package reviewhttp_test

import (
	"context"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validCourseID = "22222222-2222-2222-2222-222222222222"
const validContentID = "33333333-3333-3333-3333-333333333333"
const validArticleID = "44444444-4444-4444-4444-444444444444"

func TestCreateCourseReview(t *testing.T) {
	Convey("POST /api/v1/courses/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			createCourseReview: func(_ context.Context, _ reviewdomain.CreateCourseReviewRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/courses/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"course_id":"` + validCourseID + `","rating":5}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid course_id → 400 (request validation, before it reaches the service)", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(`{"course_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid rating → 400 (request validation, before it reaches the service)", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(`{"course_id":"`+validCourseID+`","rating":9}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("already reviewed → 422", func() {
			svcErr = reviewdomain.ErrAlreadyReviewed
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("no permission → 403", func() {
			svcErr = reviewdomain.ErrNoPermission
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Course review created successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withValidUUIDUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestCreateContentReview(t *testing.T) {
	Convey("POST /api/v1/content/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			createContentReview: func(_ context.Context, _ reviewdomain.CreateContentReviewRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/content/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"content_id":"` + validContentID + `","rating":4}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid content_id → 400 (request validation, before it reaches the service)", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(`{"content_id":"---","rating":4}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("no permission → 403", func() {
			svcErr = reviewdomain.ErrNoPermission
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Content review created successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withValidUUIDUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestCreateCourseReviewAdmin(t *testing.T) {
	Convey("POST /api/v1/admin/courses/reviews", t, func() {
		var svcErr error
		var gotAdminID string
		svc := &mockService{
			createCourseReviewAdmin: func(_ context.Context, req reviewdomain.CreateCourseReviewRequest) error {
				gotAdminID = req.UserID
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/courses/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"course_id":"` + validCourseID + `","user_id":"` + validUserID + `","rating":5}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("user_id in the body is ignored — the authenticated admin is used", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(`{"course_id":"`+validCourseID+`","user_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			So(gotAdminID, ShouldEqual, validUserID)
		})

		Convey("already reviewed → 422", func() {
			svcErr = reviewdomain.ErrAlreadyReviewed
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Course review created successfully")
		})

		Convey("Valid request → passes the authenticated user as adminID, not a user_id from the body", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldBeIn, http.StatusOK, http.StatusCreated)
			So(gotAdminID, ShouldEqual, validUserID)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withValidUUIDUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestCreateContentReviewAdmin(t *testing.T) {
	Convey("POST /api/v1/admin/content/reviews", t, func() {
		var svcErr error
		var gotAdminID string
		svc := &mockService{
			createContentReviewAdmin: func(_ context.Context, req reviewdomain.CreateContentReviewRequest) error {
				gotAdminID = req.UserID
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/content/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"content_id":"` + validContentID + `","user_id":"` + validUserID + `","rating":4}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("user_id in the body is ignored — the authenticated admin is used", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(`{"content_id":"`+validContentID+`","user_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			So(gotAdminID, ShouldEqual, validUserID)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Content review created successfully")
		})

		Convey("Valid request → passes the authenticated user as adminID, not a user_id from the body", func() {
			w := testutil.ServeHTTP(mux, withValidUUIDUser(newReq(validBody, nil)))
			So(w.Code, ShouldBeIn, http.StatusOK, http.StatusCreated)
			So(gotAdminID, ShouldEqual, validUserID)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withValidUUIDUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}
