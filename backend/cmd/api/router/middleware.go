package router

import (
	"fmt"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// RecoverPanic recovers from panics, logs the error, and returns a 500 response.
func (routes *RouteHandler) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				routes.App.Logger.Error(fmt.Errorf("panic: %v", err), map[string]any{
					"method": r.Method,
					"url":    r.URL.String(),
				})

				err = helpers.WriteJSON(w, http.StatusInternalServerError, helpers.Envelope{"error": "internal server error"}, nil)
				if err != nil {
					routes.App.Logger.Error(fmt.Errorf("response err: %v", err), nil)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// EnableCORS sets CORS headers for trusted origins and handles preflight requests.
func (routes *RouteHandler) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
		w.Header().Add("Vary", "Accept-Encoding")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range routes.App.Config.Cors.TrustedOrigins {
				if origin == routes.App.Config.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "*")
						w.Header().Set("Access-Control-Allow-Headers", "*")
						w.Header().Set("Access-Control-Max-Age", "86400")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

type rateLimitClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (route *RouteHandler) startRateLimitCleanup(mu *sync.Mutex, clients map[string]*rateLimitClient) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-route.App.Ctx.Done():
			return
		case <-ticker.C:
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}
}

// RateLimit enforces per-IP request rate limiting using a token bucket algorithm.
func (route *RouteHandler) RateLimit(next http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		clients = make(map[string]*rateLimitClient)
	)

	go route.startRateLimitCleanup(&mu, clients)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !route.App.Config.Limiter.Enabled {
			next.ServeHTTP(w, r)
			return
		}
		ip := realip.FromRequest(r)
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &rateLimitClient{
				limiter: rate.NewLimiter(rate.Limit(route.App.Config.Limiter.Rps), route.App.Config.Limiter.Burst),
			}
		}
		clients[ip].lastSeen = time.Now()
		c := clients[ip]
		allowed := c.limiter.Allow()
		mu.Unlock()
		if !allowed {
			route.RateLimitExceededResponse(w, r, c.limiter)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AuthenticateUser validates the JWT token from Authorization header and sets user in context.
func (route *RouteHandler) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// code will be added in the future when we implement authentication
		next.ServeHTTP(w, r)
	})
}
