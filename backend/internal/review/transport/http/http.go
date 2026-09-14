package reviewhttp

import (
	"net/http"

	"learnflow_backend/internal/infrastructure/logger"
	reviewdomain "learnflow_backend/internal/review/domain"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the courses module.
type Handler struct {
	svc        reviewdomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the courses module.
func NewHTTPHandler(svc reviewdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all course HTTP routes on the given mux.
// chain applies to public routes, adminChain to admin-only routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, staticChain, StaticWithAuth, adminChain alice.Chain) {

	mux.Handle("PUT /api/v1/admin/courses/reviews", adminChain.ThenFunc(h.updateCourseReviewAdmin))
	mux.Handle("PUT /api/v1/admin/content/reviews", adminChain.ThenFunc(h.updateContentReviewAdmin))

	mux.Handle("POST /api/v1/admin/courses/reviews", adminChain.ThenFunc(h.createCourseReviewAdmin))
	mux.Handle("POST /api/v1/admin/content/reviews", adminChain.ThenFunc(h.createContentReviewAdmin))

	mux.Handle("DELETE /api/v1/admin/courses/reviews/{id}", adminChain.ThenFunc(h.deleteCourseReviewAdmin))
	mux.Handle("DELETE /api/v1/admin/content/reviews/{id}", adminChain.ThenFunc(h.deleteContentReviewAdmin))

	mux.Handle("PUT /api/v1/courses/reviews", StaticWithAuth.ThenFunc(h.updateCourseReview))
	mux.Handle("PUT /api/v1/content/reviews", StaticWithAuth.ThenFunc(h.updateContentReview))

	mux.Handle("POST /api/v1/courses/reviews", StaticWithAuth.ThenFunc(h.createCourseReview))
	mux.Handle("POST /api/v1/content/reviews", StaticWithAuth.ThenFunc(h.createContentReview))

	mux.Handle("DELETE /api/v1/courses/reviews/{id}", StaticWithAuth.ThenFunc(h.deleteCourseReview))
	mux.Handle("DELETE /api/v1/content/reviews/{id}", StaticWithAuth.ThenFunc(h.deleteContentReview))

	mux.Handle("GET /api/v1/courses/{id}/reviews", staticChain.ThenFunc(h.listCourseReviews))
	mux.Handle("GET /api/v1/content/{id}/reviews", staticChain.ThenFunc(h.listContentReviews))

	mux.Handle("GET /api/v1/admin/courses/{id}/reviews", adminChain.ThenFunc(h.listCourseReviewsAdmin))
	mux.Handle("GET /api/v1/admin/content/{id}/reviews", adminChain.ThenFunc(h.listContentReviewsAdmin))

	mux.Handle("GET /api/v1/courses/{id}/reviews/stats", staticChain.ThenFunc(h.courseReviewStats))
	mux.Handle("GET /api/v1/content/{id}/reviews/stats", staticChain.ThenFunc(h.contentReviewStats))
}
