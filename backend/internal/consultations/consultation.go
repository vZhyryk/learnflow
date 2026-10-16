// Package consultations wires the consultations module: configuration and HTTP routes.
package consultations

import (
	"net/http"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	consultationshttp "learnflow_backend/internal/consultations/transport/http"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// RegisterConsultationsRoutes registers the consultations routes: userChain for clients, adminChain for admins.
func RegisterConsultationsRoutes(mux *http.ServeMux, svc consultationsdomain.Service, userChain, adminChain alice.Chain, jsonLogger *logger.Logger) {
	consultationshttp.NewHTTPHandler(svc, jsonLogger).RegisterRoutes(mux, userChain, adminChain)
}
