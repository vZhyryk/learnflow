package consultationshttp

import (
	"net/http"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the consultations module.
type Handler struct {
	svc        consultationsdomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the consultations module.
func NewHTTPHandler(svc consultationsdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all consultations HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, staticChain, adminChain alice.Chain) {
	mux.Handle("GET /v1/consultations/consultants", staticChain.ThenFunc(h.getConsultants))
	mux.Handle("GET /v1/consultations/slots", staticChain.ThenFunc(h.getAvailableSlots))
	mux.Handle("GET /v1/admin/consultations/slots", adminChain.ThenFunc(h.getAllSlots))
	mux.Handle("PATCH /v1/admin/consultations/slots", adminChain.ThenFunc(h.updateSlots))
}
