// Package router wires all HTTP routes to their handlers.
package router

import (
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"

	"github.com/justinas/alice"
)

// RouteHandler holds the compiled ServeMux and a reference to the shared App container.
type RouteHandler struct {
	Router http.Handler
	App    *app.App
}

// NewRouter registers all routes and returns a RouteHandler ready to serve.
func NewRouter(a *app.App) *RouteHandler {
	router := http.NewServeMux()
	route := &RouteHandler{
		Router: router,
		App:    a,
	}

	router.Handle("/", http.HandlerFunc(route.NotFoundResponse))

	static := alice.New(route.RecoverPanic, route.RateLimit, route.EnableCORS)
	staticWithAuth := static.Append(route.AuthenticateUser)

	// Authentication
	router.Handle("POST /api/v1/auth/login", static.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// user login logic
	}))

	router.Handle("POST /api/v1/auth/register", static.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// register a new user logic
	}))

	router.Handle("POST /api/v1/auth/logout", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// logout logic
	}))

	// Profile
	router.Handle("GET /api/v1/users/profile", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// get user profile logic
	}))
	router.Handle("PATCH /api/v1/users/profile", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// update user profile logic
	}))

	// Password
	router.Handle("POST /api/v1/users/profile/password/reset", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// request password reset logic
	}))
	router.Handle("PUT /api/v1/users/profile/password", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// set new password logic
	}))

	// Email
	router.Handle("GET /api/v1/users/profile/email/verify", static.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// verify email logic
	}))

	// Monitoring
	router.Handle("GET /metrics", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// handle prometheus metrics logic
	}))

	router.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		err := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"status": "ok"}, nil)
		if err != nil {
			route.App.Logger.Error(err, nil)
		}
	}))

	return route
}
