package courses

import (
	coursedomain "learnflow_backend/internal/courses/domain"
	coursehttp "learnflow_backend/internal/courses/transport/http"
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"

	"github.com/justinas/alice"
)

// RegisterCourseRoutes wires the courses module's HTTP handler onto mux.
func RegisterCourseRoutes(mux *http.ServeMux, svc coursedomain.Service, chain, adminChain alice.Chain, jsonLogger *logger.Logger) {
	courseHandler := coursehttp.NewHTTPHandler(svc, jsonLogger)
	courseHandler.RegisterRoutes(mux, chain, adminChain)
}
