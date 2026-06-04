// Package router wires all HTTP routes to their handlers.
package router

import (
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
	"sync"
	"time"

	"github.com/justinas/alice"
	"golang.org/x/time/rate"
)

// RouteHandler holds the compiled ServeMux and a reference to the shared App container.
type RouteHandler struct {
	Router           http.Handler
	App              *app.App
	rateLimitMu      *sync.Mutex
	rateLimitClients map[string]*rateLimitClient
}

// NewRouter registers all routes and returns a RouteHandler ready to serve.
func NewRouter(a *app.App) *RouteHandler {
	router := http.NewServeMux()
	route := &RouteHandler{
		Router:           router,
		App:              a,
		rateLimitMu:      &sync.Mutex{},
		rateLimitClients: make(map[string]*rateLimitClient),
	}

	router.Handle("/", http.HandlerFunc(route.NotFoundResponse))

	static := alice.New(route.RecoverPanic, route.RateLimit, route.EnableCORS, route.SetSecurityHeaders)
	staticWithAuth := static.Append(route.AuthenticateUser)

	if a.Config.Limiter.Enabled {
		go route.startRateLimitCleanup()
	}

	// authRepo := authrepository.NewRepository(a.DB)
	// authSvc := authservice.New(authservice.Dependencies{
	// 	Users:    authRepo,
	// 	Tokens:   authRepo,
	// 	Sessions: authRepo,
	// })
	// auth.RegisterAuthRoutes(router, authSvc)

	// Profile
	router.Handle("GET /api/v1/users/profile", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// get user profile logic
	}))
	router.Handle("PATCH /api/v1/users/profile", staticWithAuth.ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// update user profile logic
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

type rateLimitClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (route *RouteHandler) startRateLimitCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-route.App.Ctx.Done():
			return
		case <-ticker.C:
			route.rateLimitMu.Lock()
			for ip, c := range route.rateLimitClients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(route.rateLimitClients, ip)
				}
			}
			route.rateLimitMu.Unlock()
		}
	}
}
