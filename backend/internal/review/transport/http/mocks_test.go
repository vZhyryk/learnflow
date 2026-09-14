package reviewhttp_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	authdomain "learnflow_backend/internal/auth/domain"
	reviewdomain "learnflow_backend/internal/review/domain"
	reviewhttp "learnflow_backend/internal/review/transport/http"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/justinas/alice"
)

type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody
var withUser = testutil.WithUser

// validUserID is a well-formed UUID for handlers whose request Validate() checks
// UserID as a UUID (the Create* requests) — testutil.WithUser's fixed "user-123" is
// not a valid UUID and would always fail that check.
const validUserID = "11111111-1111-1111-1111-111111111111"

func withValidUUIDUser(r *http.Request) *http.Request {
	user := &authdomain.User{ID: validUserID}
	return r.WithContext(appcontext.WithUser(r.Context(), user))
}

func newAuthMux(svc *mockService) *http.ServeMux {
	h := reviewhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux, alice.Chain{}, alice.Chain{}, alice.Chain{})
	return mux
}

// httpFixture wires a mockService-backed mux and a request builder for a single
// route, shared by every per-handler fixture in this package (loginFixture,
// registerFixture, ...). Embed it and add the handler-specific svcResult/svcErr
// fields on top.
type httpFixture struct {
	mux    *http.ServeMux
	newReq func(body string, urlParams map[string]string) *http.Request
}

func newHTTPFixture(svc *mockService, method, path string) *httpFixture {
	return &httpFixture{
		mux: newAuthMux(svc),
		newReq: func(body string, urlParams map[string]string) *http.Request {
			if len(urlParams) > 0 {
				path += "?"
			}

			for key, value := range urlParams {
				if value != "" && key != "" {
					path += fmt.Sprintf("%s=%s", key, url.QueryEscape(value))
				}
			}

			return httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
		},
	}
}

type mockService struct {
	createCourseReview       func(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error
	createContentReview      func(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error
	createCourseReviewAdmin  func(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error
	createContentReviewAdmin func(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error

	updateCourseReview       func(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error
	updateContentReview      func(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error
	updateCourseReviewAdmin  func(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error
	updateContentReviewAdmin func(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error

	getCourseReviews         func(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error)
	getContentReviews        func(ctx context.Context, params pagination.Params, contentID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error)
	getCourseReviewStats     func(ctx context.Context, courseID string) (rating float64, count int, err error)
	getContentReviewStats    func(ctx context.Context, contentID string) (rating float64, count int, err error)
	deleteCourseReview       func(ctx context.Context, reviewID, userID string) error
	deleteContentReview      func(ctx context.Context, reviewID, userID string) error
	deleteCourseReviewAdmin  func(ctx context.Context, reviewID string) error
	deleteContentReviewAdmin func(ctx context.Context, reviewID string) error
}

func (m *mockService) CreateCourseReview(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error {
	if m.createCourseReview == nil {
		panic("mockService.createCourseReview not set")
	}
	return m.createCourseReview(ctx, req)
}

func (m *mockService) CreateContentReview(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error {
	if m.createContentReview == nil {
		panic("mockService.createContentReview not set")
	}
	return m.createContentReview(ctx, req)
}

func (m *mockService) CreateCourseReviewAdmin(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error {
	if m.createCourseReviewAdmin == nil {
		panic("mockService.createCourseReviewAdmin not set")
	}
	return m.createCourseReviewAdmin(ctx, req)
}

func (m *mockService) CreateContentReviewAdmin(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error {
	if m.createContentReviewAdmin == nil {
		panic("mockService.createContentReviewAdmin not set")
	}
	return m.createContentReviewAdmin(ctx, req)
}

func (m *mockService) UpdateCourseReview(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error {
	if m.updateCourseReview == nil {
		panic("mockService.updateCourseReview not set")
	}
	return m.updateCourseReview(ctx, req)
}

func (m *mockService) UpdateContentReview(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error {
	if m.updateContentReview == nil {
		panic("mockService.updateContentReview not set")
	}
	return m.updateContentReview(ctx, req)
}

func (m *mockService) UpdateCourseReviewAdmin(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error {
	if m.updateCourseReviewAdmin == nil {
		panic("mockService.updateCourseReviewAdmin not set")
	}
	return m.updateCourseReviewAdmin(ctx, req)
}

func (m *mockService) UpdateContentReviewAdmin(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error {
	if m.updateContentReviewAdmin == nil {
		panic("mockService.updateContentReviewAdmin not set")
	}
	return m.updateContentReviewAdmin(ctx, req)
}

func (m *mockService) GetCourseReviews(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
	if m.getCourseReviews == nil {
		panic("mockService.getCourseReviews not set")
	}
	return m.getCourseReviews(ctx, params, courseID, filter)
}

func (m *mockService) GetContentReviews(ctx context.Context, params pagination.Params, contentID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
	if m.getContentReviews == nil {
		panic("mockService.getContentReviews not set")
	}
	return m.getContentReviews(ctx, params, contentID, filter)
}

func (m *mockService) GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error) {
	if m.getCourseReviewStats == nil {
		panic("mockService.getCourseReviewStats not set")
	}
	return m.getCourseReviewStats(ctx, courseID)
}

func (m *mockService) GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error) {
	if m.getContentReviewStats == nil {
		panic("mockService.getContentReviewStats not set")
	}
	return m.getContentReviewStats(ctx, contentID)
}

func (m *mockService) DeleteCourseReview(ctx context.Context, reviewID, userID string) error {
	if m.deleteCourseReview == nil {
		panic("mockService.deleteCourseReview not set")
	}
	return m.deleteCourseReview(ctx, reviewID, userID)
}

func (m *mockService) DeleteContentReview(ctx context.Context, reviewID, userID string) error {
	if m.deleteContentReview == nil {
		panic("mockService.deleteContentReview not set")
	}
	return m.deleteContentReview(ctx, reviewID, userID)
}

func (m *mockService) DeleteCourseReviewAdmin(ctx context.Context, reviewID string) error {
	if m.deleteCourseReviewAdmin == nil {
		panic("mockService.deleteCourseReviewAdmin not set")
	}
	return m.deleteCourseReviewAdmin(ctx, reviewID)
}

func (m *mockService) DeleteContentReviewAdmin(ctx context.Context, reviewID string) error {
	if m.deleteContentReviewAdmin == nil {
		panic("mockService.deleteContentReviewAdmin not set")
	}
	return m.deleteContentReviewAdmin(ctx, reviewID)
}
