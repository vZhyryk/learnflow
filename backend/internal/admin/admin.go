package admin

import (
	admindomain "learnflow_backend/internal/admin/domain"
	adminhttp "learnflow_backend/internal/admin/transport/http"
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"

	"github.com/justinas/alice"
)

// RegisterAdminRoutes wires the admin module's HTTP handler onto mux.
func RegisterAdminRoutes(mux *http.ServeMux, svc admindomain.Service, adminChain, staticChain alice.Chain, jsonLogger *logger.Logger) {
	adminHandler := adminhttp.NewHTTPHandler(svc, jsonLogger)
	adminHandler.RegisterRoutes(mux, adminChain, staticChain)
}
