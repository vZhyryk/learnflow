package auth

import (
	"net/http"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"
)

// RegisterAuthRoutes wires all auth HTTP routes onto mux.
func RegisterAuthRoutes(mux *http.ServeMux, svc authdomain.Service) {
	authHandler := authhttp.NewHTTPHandler(svc)
	authHandler.RegisterRoutes(mux)
}
