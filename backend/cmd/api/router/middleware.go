package router

import (
	"context"
	"fmt"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

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
					"stack":  string(debug.Stack()),
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
		w.Header().Add("Vary", "Cookie")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
		w.Header().Add("Vary", "Accept-Encoding")

		origin := r.Header.Get("Origin")

		if origin != "" {
			if _, ok := routes.App.Config.Cors.TrustedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					w.Header().Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// SetSecurityHeaders sets basic HTTP security headers on every response.
func (routes *RouteHandler) SetSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'; script-src 'self' 'wasm-unsafe-eval'")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-store")
		next.ServeHTTP(w, r)
	})
}

// NewRouteRateLimiter creates a rate limiter middleware using the provided key function and request/time budget.
func (route *RouteHandler) NewRouteRateLimiter(reqCount float64, duration time.Duration, burst int, getKeyFunc func(*http.Request) string) func(next http.Handler) http.Handler {
	rps := reqCount / duration.Seconds()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !route.App.Config.Limiter.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			key := getKeyFunc(r)

			route.rateLimitMu.Lock()
			if _, found := route.rateLimitClients[key]; !found {
				route.rateLimitClients[key] = &rateLimitClient{
					limiter: rate.NewLimiter(rate.Limit(rps), burst),
				}
			}

			route.rateLimitClients[key].lastSeen = time.Now()
			allowed := route.rateLimitClients[key].limiter.Allow()
			route.rateLimitMu.Unlock()
			if !allowed {
				route.RateLimitExceededResponse(w, route.rateLimitClients[key].limiter)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout enforces a 10-second request timeout by wrapping the context.
func (route *RouteHandler) Timeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthenticateUser validates the JWT token from Authorization header and sets user in context.
func (route *RouteHandler) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// code will be added in the future when we implement authentication
		next.ServeHTTP(w, r)
	})
}

func realClientIP(r *http.Request, trustedProxies []net.IPNet) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	ip := net.ParseIP(remoteIP)

	for _, cidr := range trustedProxies {
		if cidr.Contains(ip) {
			if client := ipFromProxyHeaders(r); client != "" {
				return client
			}
			break
		}
	}
	return remoteIP
}

func ipFromProxyHeaders(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := parseIP(strings.Split(xff, ",")[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return parseIP(xri)
	}
	return ""
}

func parseIP(s string) string {
	if parsed := net.ParseIP(strings.TrimSpace(s)); parsed != nil {
		return parsed.String()
	}
	return ""
}

// SetIPAddress extracts the client IP from headers and stores it in context.
func (routes *RouteHandler) SetIPAddress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realClientIP(r, routes.App.Config.TrustedProxies)
		ctx := appcontext.WithIPAddress(r.Context(), ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

const requestIDHeader = "X-Request-ID"

// SetRequestID generates or retrieves a request ID and stores it in context and response headers.
func (routes *RouteHandler) SetRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = appcontext.NewRequestID()
		}

		w.Header().Set(requestIDHeader, requestID)
		ctx := appcontext.WithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
