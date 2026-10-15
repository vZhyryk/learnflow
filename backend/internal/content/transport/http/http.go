package contenthttp

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"net/http"

	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the content module.
type Handler struct {
	svc        contentdomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the content module.
func NewHTTPHandler(svc contentdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all content HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, staticChain, adminChain alice.Chain) {
	mux.Handle("GET /api/v1/content", staticChain.ThenFunc(h.listContentItems))
	mux.Handle("GET /api/v1/content/{slug}", staticChain.ThenFunc(h.getContentItemBySlug))

	mux.Handle("GET /api/v1/admin/content", adminChain.ThenFunc(h.listAllContentItems))
	mux.Handle("POST /api/v1/admin/content", adminChain.ThenFunc(h.createContentItem))
	mux.Handle("PUT /api/v1/admin/content", adminChain.ThenFunc(h.updateContentItem))
	mux.Handle("PUT /api/v1/admin/content/{id}/publish", adminChain.ThenFunc(h.publishContentItem))
	mux.Handle("PUT /api/v1/admin/content/{id}/archive", adminChain.ThenFunc(h.archiveContentItem))
	mux.Handle("DELETE /api/v1/admin/content/{id}", adminChain.ThenFunc(h.deleteContentItem))
}
