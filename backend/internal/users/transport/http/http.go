package usershttp

import (
	"learnflow_backend/internal/infrastructure/logger"
	usersdomain "learnflow_backend/internal/users/domain"
	"net/http"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the users module.
type Handler struct {
	svc        usersdomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the users module.
func NewHTTPHandler(svc usersdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all user HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, chain alice.Chain) {
	mux.Handle("GET /api/v1/users/profile", chain.ThenFunc(h.getProfile))

	mux.Handle("PATCH /api/v1/users/profile", chain.ThenFunc(h.changeProfile))
}
