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
func (h *Handler) RegisterRoutes(mux *http.ServeMux, adminChain, staticChain alice.Chain) {
	mux.Handle("GET /api/v1/announcements", staticChain.ThenFunc(h.getPublicAnnouncements))

	mux.Handle("POST /api/v1/admin/announcements", adminChain.ThenFunc(h.createAnnouncement))
	mux.Handle("PUT /api/v1/admin/announcements", adminChain.ThenFunc(h.updateAnnouncement))
	mux.Handle("PUT /api/v1/admin/announcements/{id}/approve", adminChain.ThenFunc(h.approveAnnouncement))

	mux.Handle("GET /api/v1/admin/announcements/all", adminChain.ThenFunc(h.getAnnouncements))
	mux.Handle("GET /api/v1/admin/announcements/unapproved", adminChain.ThenFunc(h.getUnApprovedAnnouncements))
	mux.Handle("GET /api/v1/admin/announcements/approved", adminChain.ThenFunc(h.getApprovedAnnouncements))
	mux.Handle("GET /api/v1/admin/announcements/expired", adminChain.ThenFunc(h.getExpiredAnnouncements))

	mux.Handle("GET /api/v1/admin/users", adminChain.ThenFunc(h.getUsersData))
	mux.Handle("GET /api/v1/admin/users/{id}", adminChain.ThenFunc(h.getUserDataByID))
	mux.Handle("DELETE /api/v1/admin/users/{id}", adminChain.ThenFunc(h.deleteUser))
	mux.Handle("PUT /api/v1/admin/users/{id}/restore", adminChain.ThenFunc(h.restoreUser))
	mux.Handle("PUT /api/v1/admin/users/{id}/block", adminChain.ThenFunc(h.blockUser))
	mux.Handle("PUT /api/v1/admin/users/{id}/unblock", adminChain.ThenFunc(h.unBlockUser))
	mux.Handle("PUT /api/v1/admin/users/{id}/subadmin", adminChain.ThenFunc(h.assignUserRole))
	mux.Handle("DELETE /api/v1/admin/users/{id}/subadmin", adminChain.ThenFunc(h.revokeUserRole))
}
