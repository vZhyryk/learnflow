package content

import (
	contentdomain "learnflow_backend/internal/content/domain"
	contenthttp "learnflow_backend/internal/content/transport/http"
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"

	"github.com/justinas/alice"
)

// RegisterContentRoutes wires the content module's HTTP handler onto mux.
func RegisterContentRoutes(mux *http.ServeMux, svc contentdomain.Service, staticChain, adminChain alice.Chain, jsonLogger *logger.Logger) {
	contentHandler := contenthttp.NewHTTPHandler(svc, jsonLogger)
	contentHandler.RegisterRoutes(mux, staticChain, adminChain)
}
