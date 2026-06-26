package users

import (
	"net/http"

	"learnflow_backend/internal/infrastructure/logger"
	usersdomain "learnflow_backend/internal/users/domain"
	usershttp "learnflow_backend/internal/users/transport/http"

	"github.com/justinas/alice"
)

// RegisterUsersRoutes wires the users HTTP handler into the application mux.
func RegisterUsersRoutes(mux *http.ServeMux, svc usersdomain.Service, chain alice.Chain, jsonLogger *logger.Logger) {
	usersHandler := usershttp.NewHTTPHandler(svc, jsonLogger)
	usersHandler.RegisterRoutes(mux, chain)
}
