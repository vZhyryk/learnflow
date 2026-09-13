package review

import (
	"learnflow_backend/internal/infrastructure/logger"
	reviewservice "learnflow_backend/internal/review/service"
	reviewhttp "learnflow_backend/internal/review/transport/http"
	"net/http"

	"github.com/justinas/alice"
)

// RegisterReviewRoutes wires the review module's HTTP handler onto mux.
func RegisterReviewRoutes(mux *http.ServeMux, svc *reviewservice.Service, staticChain, staticWithAuth, adminChain alice.Chain, jsonLogger *logger.Logger) {
	reviewHandler := reviewhttp.NewHTTPHandler(svc, jsonLogger)
	reviewHandler.RegisterRoutes(mux, staticChain, staticWithAuth, adminChain)
}
