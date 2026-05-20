// Package router wires all HTTP routes to their handlers.
package router

import (
	"learnflow_backend/cmd/api/app"
	"net/http"
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

	router.HandleFunc("POST /api/v1/resource", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return route
}
