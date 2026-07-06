// Package router wires all HTTP routes to their handlers.
package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/internal/auth"
	authdomain "learnflow_backend/internal/auth/domain"
	authrepository "learnflow_backend/internal/auth/repository"
	authservice "learnflow_backend/internal/auth/service"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/tokens"
	"learnflow_backend/internal/users"
	usersrepository "learnflow_backend/internal/users/repository"
	usersservice "learnflow_backend/internal/users/service"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RouteHandler holds the compiled ServeMux and a reference to the shared App container.
type RouteHandler struct {
	Router http.Handler
	App    *app.App
	token  *tokens.Tokens
}

// NewRouter registers all routes and returns a RouteHandler ready to serve.
func NewRouter(a *app.App) (*RouteHandler, error) {
	router := http.NewServeMux()
	route := &RouteHandler{
		Router: router,
		App:    a,
		token:  tokens.NewTokens(a.Config.Secret.JWTSecret, a.Config.Secret.JWTSecretPrev, a.Config.Secret.JWTIssuer, a.Config.Secret.JWTAudience),
	}

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := helpers.NotFoundResponse(w); err != nil {
			a.Logger.Error(err, nil)
		}
	}))

	chains := route.buildChains()
	authRepo := authrepository.NewRepository(a.DB)
	transactor := db.NewTransactor(a.DB)
	outbox := events.NewOutboxWriter(a.DB)
	authSvc, err := authservice.New(
		authservice.Repos{
			UserRepo:    authRepo,
			SessionRepo: authRepo,
			TokenRepo:   authRepo,
			Transactor:  transactor,
		},
		authservice.Utils{
			Outbox:      outbox,
			Token:       route.token,
			RedisClient: a.Redis,
		},
		authservice.Options{})
	if err != nil {
		return nil, fmt.Errorf("router: NewRouter: %w", err)
	}

	auth.RegisterAuthRoutes(router, authSvc, chains, a.Logger)

	usersrepo := usersrepository.NewRepository(a.DB)
	userSvc := usersservice.New(usersrepo)
	users.RegisterUsersRoutes(router, userSvc, chains.StaticWithAuth, a.Logger)

	router.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"status": "ok"}, nil)
		if err != nil {
			route.App.Logger.Error(err, map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
			})
		}
	}))

	router.Handle("GET /readiness", http.HandlerFunc(route.Readiness))

	router.Handle("GET /metrics", promhttp.Handler())

	return route, nil
}

const rateLimitBodyLimit = 4_096

// bodyRateLimitKey reads a bounded JSON body and passes the raw bytes to extractField to pull
// out the value to key the rate limiter on, then resets r.Body so downstream handlers
// (ReadJSON) receive the full body. Do NOT add any body-reading middleware between this
// call and the handler.
func (route *RouteHandler) bodyRateLimitKey(r *http.Request, extractField func(bodyBytes []byte) (value string, ok bool)) string {
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, rateLimitBodyLimit))
	if err != nil {
		return ""
	}

	if err := r.Body.Close(); err != nil {
		route.App.Logger.Error(err, nil)
	}

	if len(bodyBytes) >= rateLimitBodyLimit {
		return "oversized-body"
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	value, ok := extractField(bodyBytes)
	if !ok || value == "" {
		return appcontext.IPAddressFromContext(r.Context())
	}

	return tokens.MakeHash(value)
}

func (route *RouteHandler) getEmailFromBody(r *http.Request) string {
	return route.bodyRateLimitKey(r, func(bodyBytes []byte) (string, bool) {
		var req authdomain.RequestPasswordResetRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			return "", false
		}
		return req.Email, true
	})
}

func (route *RouteHandler) getTokenFromBody(r *http.Request) string {
	return route.bodyRateLimitKey(r, func(bodyBytes []byte) (string, bool) {
		var req authdomain.VerifyEmailRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			return "", false
		}
		return req.Token, true
	})
}

func (route *RouteHandler) buildChains() authhttp.AuthRouteChains {
	staticLimiter := route.NewRouteRateLimiter(route.App.Config.Limiter.Rps, time.Second, route.App.Config.Limiter.Burst, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context())
	})

	staticChain := route.SetChain(staticLimiter)

	loginLimiter := route.NewRouteRateLimiter(5, time.Minute, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + route.getEmailFromBody(r)
	})

	registerLimiter := route.NewRouteRateLimiter(3, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context())
	})

	passResetLimiter := route.NewRouteRateLimiter(2, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + route.getEmailFromBody(r)
	})

	emailVerifyLimiter := route.NewRouteRateLimiter(3, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + route.getTokenFromBody(r)
	})

	return authhttp.AuthRouteChains{
		Static:         staticChain,
		Login:          route.SetChain(loginLimiter),
		Register:       route.SetChain(registerLimiter),
		PassReset:      route.SetChain(passResetLimiter),
		EmailVerify:    route.SetChain(emailVerifyLimiter),
		StaticWithAuth: staticChain.Append(route.AuthenticateUser),
	}
}

// Readiness checks DB and Redis connectivity and returns 200 if the service is ready to handle traffic.
func (h *RouteHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.App.DB.Ping(ctx); err != nil {
		if respErr := helpers.WriteJSON(w, http.StatusServiceUnavailable, helpers.Envelope{"status": "unavailable", "reason": "database"}, nil); respErr != nil {
			h.App.Logger.Error(respErr, map[string]any{
				"status":  http.StatusServiceUnavailable,
				"envelop": helpers.Envelope{"status": "unavailable", "reason": "database"},
			})
		}
		return
	}
	if err := h.App.Redis.Ping(ctx).Err(); err != nil {
		if respErr := helpers.WriteJSON(w, http.StatusServiceUnavailable, helpers.Envelope{"status": "unavailable", "reason": "redis"}, nil); respErr != nil {
			h.App.Logger.Error(respErr, map[string]any{
				"status":  http.StatusServiceUnavailable,
				"envelop": helpers.Envelope{"status": "unavailable", "reason": "redis"},
			})
		}
		return
	}
	if respErr := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"status": "ready"}, nil); respErr != nil {
		h.App.Logger.Error(respErr, map[string]any{
			"status":  http.StatusOK,
			"envelop": helpers.Envelope{"status": "ready"},
		})
	}
}

// SetChain builds the standard middleware chain, inserting limiter for rate limiting.
func (route *RouteHandler) SetChain(limiter func(http.Handler) http.Handler) alice.Chain {
	return alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger, limiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders)
}
