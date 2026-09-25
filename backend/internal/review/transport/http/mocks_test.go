package reviewhttp_test

import (
	"context"
	"net/http"

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
	f := testutil.NewHTTPFixture(newAuthMux(svc), method, path)
	return &httpFixture{mux: f.Mux, newReq: f.NewReq}
}

type mockService struct {
	createCourseReview       func(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error
	createContentReview      func(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error
	createCourseReviewAdmin  func(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error
	createContentReviewAdmin func(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error

	updateCourseReview       func(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error
	updateContentReview      func(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error
	updateCourseReviewAdmin  func(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest, adminID string) error
	updateContentReviewAdmin func(ctx context.Context, req reviewdomain.UpdateContentReviewRequest, adminID string) error

	getCourseReviews         func(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error)
	getContentReviews        func(ctx context.Context, params pagination.Params, contentID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error)
	getCourseReviewStats     func(ctx context.Context, courseID string) (rating float64, count int, err error)
	getContentReviewStats    func(ctx context.Context, contentID string) (rating float64, count int, err error)
	deleteCourseReview       func(ctx context.Context, reviewID, userID string) error
	deleteContentReview      func(ctx context.Context, reviewID, userID string) error
	deleteCourseReviewAdmin  func(ctx context.Context, reviewID, userID string) error
	deleteContentReviewAdmin func(ctx context.Context, reviewID, userID string) error

	createArticleReview      func(ctx context.Context, req reviewdomain.CreateArticleReviewRequest) error
	createArticleReviewAdmin func(ctx context.Context, req reviewdomain.CreateArticleReviewRequest) error
	updateArticleReview      func(ctx context.Context, req reviewdomain.UpdateArticleReviewRequest) error
	updateArticleReviewAdmin func(ctx context.Context, req reviewdomain.UpdateArticleReviewRequest, adminID string) error
	getArticleReviews        func(ctx context.Context, params pagination.Params, articleID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error)
	getArticleReviewStats    func(ctx context.Context, articleID string) (rating float64, count int, err error)
	deleteArticleReview      func(ctx context.Context, reviewID, userID string) error
	deleteArticleReviewAdmin func(ctx context.Context, reviewID, userID string) error
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

func (m *mockService) UpdateCourseReviewAdmin(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest, adminID string) error {
	if m.updateCourseReviewAdmin == nil {
		panic("mockService.updateCourseReviewAdmin not set")
	}
	return m.updateCourseReviewAdmin(ctx, req, adminID)
}

func (m *mockService) UpdateContentReviewAdmin(ctx context.Context, req reviewdomain.UpdateContentReviewRequest, adminID string) error {
	if m.updateContentReviewAdmin == nil {
		panic("mockService.updateContentReviewAdmin not set")
	}
	return m.updateContentReviewAdmin(ctx, req, adminID)
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

func (m *mockService) DeleteCourseReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if m.deleteCourseReviewAdmin == nil {
		panic("mockService.deleteCourseReviewAdmin not set")
	}
	return m.deleteCourseReviewAdmin(ctx, reviewID, userID)
}

func (m *mockService) DeleteContentReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if m.deleteContentReviewAdmin == nil {
		panic("mockService.deleteContentReviewAdmin not set")
	}
	return m.deleteContentReviewAdmin(ctx, reviewID, userID)
}

func (m *mockService) CreateArticleReview(ctx context.Context, req reviewdomain.CreateArticleReviewRequest) error {
	if m.createArticleReview == nil {
		panic("mockService.createArticleReview not set")
	}
	return m.createArticleReview(ctx, req)
}

func (m *mockService) CreateArticleReviewAdmin(ctx context.Context, req reviewdomain.CreateArticleReviewRequest) error {
	if m.createArticleReviewAdmin == nil {
		panic("mockService.createArticleReviewAdmin not set")
	}
	return m.createArticleReviewAdmin(ctx, req)
}

func (m *mockService) UpdateArticleReview(ctx context.Context, req reviewdomain.UpdateArticleReviewRequest) error {
	if m.updateArticleReview == nil {
		panic("mockService.updateArticleReview not set")
	}
	return m.updateArticleReview(ctx, req)
}

func (m *mockService) UpdateArticleReviewAdmin(ctx context.Context, req reviewdomain.UpdateArticleReviewRequest, adminID string) error {
	if m.updateArticleReviewAdmin == nil {
		panic("mockService.updateArticleReviewAdmin not set")
	}
	return m.updateArticleReviewAdmin(ctx, req, adminID)
}

func (m *mockService) GetArticleReviews(ctx context.Context, params pagination.Params, articleID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error) {
	if m.getArticleReviews == nil {
		panic("mockService.getArticleReviews not set")
	}
	return m.getArticleReviews(ctx, params, articleID, filter)
}

func (m *mockService) GetArticleReviewStats(ctx context.Context, articleID string) (rating float64, count int, err error) {
	if m.getArticleReviewStats == nil {
		panic("mockService.getArticleReviewStats not set")
	}
	return m.getArticleReviewStats(ctx, articleID)
}

func (m *mockService) DeleteArticleReview(ctx context.Context, reviewID, userID string) error {
	if m.deleteArticleReview == nil {
		panic("mockService.deleteArticleReview not set")
	}
	return m.deleteArticleReview(ctx, reviewID, userID)
}

func (m *mockService) DeleteArticleReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if m.deleteArticleReviewAdmin == nil {
		panic("mockService.deleteArticleReviewAdmin not set")
	}
	return m.deleteArticleReviewAdmin(ctx, reviewID, userID)
}
