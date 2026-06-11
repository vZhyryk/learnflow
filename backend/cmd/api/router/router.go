// Package router wires all HTTP routes to their handlers.
package router

import (
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/internal/auth"
	authrepository "learnflow_backend/internal/auth/repository"
	authservice "learnflow_backend/internal/auth/service"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
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

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := helpers.NotFoundResponse(w); err != nil {
			a.Logger.Error(err, nil)
		}
	}))

	staticLimiter := route.NewRouteRateLimiter(a.Config.Limiter.Rps, time.Second, a.Config.Limiter.Burst, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context())
	})

	staticChain := alice.New(route.RecoverPanic, route.SetIPAddress, staticLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders, route.SetRequestID)

	loginLimiter := route.NewRouteRateLimiter(5, time.Minute, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + r.FormValue("email")
	})

	loginChain := alice.New(route.RecoverPanic, route.SetIPAddress, loginLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders, route.SetRequestID)

	registerLimiter := route.NewRouteRateLimiter(3, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context())
	})
	registerChain := alice.New(route.RecoverPanic, route.SetIPAddress, registerLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders, route.SetRequestID)

	passResetLimiter := route.NewRouteRateLimiter(2, time.Hour, 1, func(r *http.Request) string {
		return r.FormValue("email")
	})
	passResetChain := alice.New(route.RecoverPanic, route.SetIPAddress, passResetLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders, route.SetRequestID)

	emailVerifyLimiter := route.NewRouteRateLimiter(3, time.Hour, 1, func(r *http.Request) string {
		return r.FormValue("email")
	})
	emailVerifyChain := alice.New(route.RecoverPanic, route.SetIPAddress, emailVerifyLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders, route.SetRequestID)

	staticWithAuth := staticChain.Append(route.AuthenticateUser)

	if a.Config.Limiter.Enabled {
		a.Wg.Add(1)
		go route.startRateLimitCleanup()
	}

	authRepo := authrepository.NewRepository(a.DB)
	transactor := db.NewTransactor(a.DB)
	authSvc := authservice.New(authRepo, authRepo, authRepo, transactor, a.Config.JWTSecret, a.Logger)
	auth.RegisterAuthRoutes(router, authSvc, authhttp.AuthRouteChains{
		Static:         staticChain,
		Login:          loginChain,
		Register:       registerChain,
		PassReset:      passResetChain,
		EmailVerify:    emailVerifyChain,
		StaticWithAuth: staticWithAuth,
	}, a.Logger)

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

	router.Handle("GET /readiness", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// викликає `db.PingContext` + Redis PING. Повертає `503 {"status":"unavailable","reason":"..."}` при збої.
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
	defer route.App.Wg.Done()
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
