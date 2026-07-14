package auth

import (
	"net/http"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/infrastructure/logger"
)

// RegisterAuthRoutes wires all auth HTTP routes onto mux.
func RegisterAuthRoutes(mux *http.ServeMux, svc authdomain.Service, chains authhttp.AuthRouteChains, jsonLogger *logger.Logger) {
	authHandler := authhttp.NewHTTPHandler(svc, jsonLogger)
	authHandler.RegisterRoutes(mux, chains)
}
