package coursehttp

import (
	"net/http"

	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the courses module.
type Handler struct {
	svc        coursedomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the courses module.
func NewHTTPHandler(svc coursedomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all course HTTP routes on the given mux.
// chain applies to public routes, adminChain to admin-only routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, chain, adminChain alice.Chain) {
	mux.Handle("GET /api/v1/courses", chain.ThenFunc(h.listCourses))
	mux.Handle("GET /api/v1/courses/{slug}", chain.ThenFunc(h.getCourseBySlug))

	mux.Handle("GET /api/v1/admin/courses", adminChain.ThenFunc(h.listAllCourses))
	mux.Handle("POST /api/v1/admin/courses", adminChain.ThenFunc(h.createCourse))
	mux.Handle("PUT /api/v1/admin/courses", adminChain.ThenFunc(h.updateCourse))
	mux.Handle("PUT /api/v1/admin/courses/{id}/publish", adminChain.ThenFunc(h.publishCourse))
	mux.Handle("PUT /api/v1/admin/courses/{id}/archive", adminChain.ThenFunc(h.archiveCourse))
	mux.Handle("DELETE /api/v1/admin/courses/{id}", adminChain.ThenFunc(h.deleteCourse))
}
