package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"

	"github.com/justinas/alice"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	svc        authdomain.Service
	jsonLogger *logger.Logger
}

// AuthRouteChains holds pre-configured middleware chains for different auth routes.
type AuthRouteChains struct {
	Static         alice.Chain
	Login          alice.Chain
	Register       alice.Chain
	PassReset      alice.Chain
	EmailVerify    alice.Chain
	StaticWithAuth alice.Chain
}

// NewHTTPHandler returns a new auth HTTP Handler.
func NewHTTPHandler(svc authdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes mounts all auth endpoints onto the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, chains AuthRouteChains) {
	mux.Handle("POST /api/v1/auth/login", chains.Login.ThenFunc(h.login))

	mux.Handle("POST /api/v1/auth/register", chains.Register.ThenFunc(h.register))

	mux.Handle("POST /api/v1/auth/logout", chains.StaticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// logout logic
	}))

	mux.Handle("POST /api/v1/users/auth/password/reset", chains.PassReset.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// request password reset logic
	}))

	mux.Handle("PUT /api/v1/users/auth/password", chains.StaticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// set new password logic
	}))

	mux.Handle("GET /api/v1/users/auth/email/verify", chains.EmailVerify.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// verify email logic
	}))

	mux.Handle("POST /api/v1/auth/refresh", chains.StaticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// refresh token
	}))
}
