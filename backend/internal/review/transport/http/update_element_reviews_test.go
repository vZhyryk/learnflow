package reviewhttp_test

import (
	"context"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validReviewID = "44444444-4444-4444-4444-444444444444"

func TestUpdateCourseReview(t *testing.T) {
	Convey("PUT /api/v1/courses/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			updateCourseReview: func(_ context.Context, _ reviewdomain.UpdateCourseReviewRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/courses/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"review_id":"` + validReviewID + `","rating":5}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid review_id → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid rating → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"`+validReviewID+`","rating":9}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("no permission → 403", func() {
			svcErr = reviewdomain.ErrNoPermission
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Course review updated successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestUpdateContentReview(t *testing.T) {
	Convey("PUT /api/v1/content/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			updateContentReview: func(_ context.Context, _ reviewdomain.UpdateContentReviewRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/content/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"review_id":"` + validReviewID + `","rating":4}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid review_id → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"---","rating":4}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("no permission → 403", func() {
			svcErr = reviewdomain.ErrNoPermission
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Content review updated successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestUpdateCourseReviewAdmin(t *testing.T) {
	Convey("PUT /api/v1/admin/courses/reviews", t, func() {
		var svcErr error
		var gotAdminID string
		svc := &mockService{
			updateCourseReviewAdmin: func(_ context.Context, req reviewdomain.UpdateCourseReviewRequest) error {
				gotAdminID = req.UserID
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/courses/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"review_id":"` + validReviewID + `","rating":5}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid review_id → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Course review updated successfully")
		})

		Convey("Valid request → passes the authenticated user as adminID, not a user_id from the body", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldBeIn, http.StatusOK, http.StatusCreated)
			So(gotAdminID, ShouldEqual, "user-123")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestUpdateContentReviewAdmin(t *testing.T) {
	Convey("PUT /api/v1/admin/content/reviews", t, func() {
		var svcErr error
		var gotAdminID string
		svc := &mockService{
			updateContentReviewAdmin: func(_ context.Context, req reviewdomain.UpdateContentReviewRequest) error {
				gotAdminID = req.UserID
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/content/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"review_id":"` + validReviewID + `","rating":4}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid review_id → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"---","rating":4}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Content review updated successfully")
		})

		Convey("Valid request → passes the authenticated user as adminID, not a user_id from the body", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldBeIn, http.StatusOK, http.StatusCreated)
			So(gotAdminID, ShouldEqual, "user-123")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestUpdateArticleReview(t *testing.T) {
	Convey("PUT /api/v1/articles/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			updateArticleReview: func(_ context.Context, _ reviewdomain.UpdateArticleReviewRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/articles/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"review_id":"` + validReviewID + `","rating":5}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid review_id → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Article review updated successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestUpdateArticleReviewAdmin(t *testing.T) {
	Convey("PUT /api/v1/admin/articles/reviews", t, func() {
		var svcErr error
		svc := &mockService{
			updateArticleReviewAdmin: func(_ context.Context, _ reviewdomain.UpdateArticleReviewRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/articles/reviews")
		mux, newReq := f.mux, f.newReq
		validBody := `{"review_id":"` + validReviewID + `","rating":5}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("invalid review_id → 400 (request validation)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"review_id":"---","rating":5}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("review not found → 404", func() {
			svcErr = reviewdomain.ErrReviewNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldEqual, "Article review updated successfully")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}
