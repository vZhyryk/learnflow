package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"net/http"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	svc authdomain.Service
}

// NewHTTPHandler returns a new auth HTTP Handler.
func NewHTTPHandler(svc authdomain.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {

	// mux.Handle("POST /api/v1/auth/login", static.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
	// 	// user login logic
	// }))

	// mux.Handle("POST /api/v1/auth/register", static.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
	// 	// register a new user logic
	// }))

	// mux.Handle("POST /api/v1/auth/logout", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
	// 	// logout logic
	// }))

	// mux.Handle("POST /api/v1/users/auth/password/reset", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
	// 	// request password reset logic
	// }))

	// mux.Handle("PUT /api/v1/users/auth/password", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
	// 	// set new password logic
	// }))

	// mux.Handle("GET /api/v1/users/auth/email/verify", static.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
	// 	// verify email logic
	// }))

}
