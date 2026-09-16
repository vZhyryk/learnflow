package adminhttp

import (
	"net/http"

	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the admin module.
type Handler struct {
	svc        admindomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the admin module.
func NewHTTPHandler(svc admindomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all admin HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, adminChain alice.Chain) {
	mux.Handle("POST /api/v1/admin/announcement", adminChain.ThenFunc(h.createAnnouncement))
	mux.Handle("PUT /api/v1/admin/announcement", adminChain.ThenFunc(h.updateAnnouncement))
	mux.Handle("PUT /api/v1/admin/announcement/{id}/approve", adminChain.ThenFunc(h.approveAnnouncement))

	mux.Handle("GET /api/v1/admin/announcement/all", adminChain.ThenFunc(h.getAnnouncements))
	mux.Handle("GET /api/v1/admin/announcement/unapproved", adminChain.ThenFunc(h.getUnApprovedAnnouncements))
	mux.Handle("GET /api/v1/admin/announcement/approved", adminChain.ThenFunc(h.getApprovedAnnouncements))
	mux.Handle("GET /api/v1/admin/announcement/expired", adminChain.ThenFunc(h.getExpiredAnnouncements))
}
