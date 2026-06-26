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
	authSvc, err := authservice.New(authRepo, authRepo, authRepo, transactor, outbox, route.token, a.Logger, a.Redis)
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

// getEmailFromBody reads the request body to extract an email for rate-limit keying,
// then resets r.Body so downstream handlers (ReadJSON) receive the full body.
// Do NOT add any body-reading middleware between this call and the handler.
func (route *RouteHandler) getEmailFromBody(r *http.Request) string {
	var req authdomain.RequestPasswordResetRequest
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 4_096))
	if err != nil {
		return ""
	}

	if err := r.Body.Close(); err != nil {
		route.App.Logger.Error(err, nil)
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return appcontext.IPAddressFromContext(r.Context())
	}

	if req.Email == "" {
		return appcontext.IPAddressFromContext(r.Context())
	}

	return tokens.MakeHash(req.Email)
}

func (route *RouteHandler) getTokenFromBody(r *http.Request) string {
	var req authdomain.VerifyEmailRequest
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 4_096))
	if err != nil {
		return ""
	}

	if err := r.Body.Close(); err != nil {
		route.App.Logger.Error(err, nil)
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return appcontext.IPAddressFromContext(r.Context())
	}

	if req.Token == "" {
		return appcontext.IPAddressFromContext(r.Context())
	}

	return tokens.MakeHash(req.Token)
}

func (route *RouteHandler) buildChains() authhttp.AuthRouteChains {
	staticLimiter := route.NewRouteRateLimiter(route.App.Config.Limiter.Rps, time.Second, route.App.Config.Limiter.Burst, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context())
	})

	staticChain := alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger, staticLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders)

	loginLimiter := route.NewRouteRateLimiter(5, time.Minute, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + route.getEmailFromBody(r)
	})

	loginChain := alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger, loginLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders)

	registerLimiter := route.NewRouteRateLimiter(3, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context())
	})

	registerChain := alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger, registerLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders)

	passResetLimiter := route.NewRouteRateLimiter(2, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + route.getEmailFromBody(r)
	})
	passResetChain := alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger, passResetLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders)

	emailVerifyLimiter := route.NewRouteRateLimiter(3, time.Hour, 1, func(r *http.Request) string {
		return appcontext.IPAddressFromContext(r.Context()) + ":" + route.getTokenFromBody(r)
	})

	emailVerifyChain := alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger, emailVerifyLimiter, route.Timeout, route.EnableCORS, route.SetSecurityHeaders)

	staticWithAuth := staticChain.Append(route.AuthenticateUser)

	return authhttp.AuthRouteChains{
		Static:         staticChain,
		Login:          loginChain,
		Register:       registerChain,
		PassReset:      passResetChain,
		EmailVerify:    emailVerifyChain,
		StaticWithAuth: staticWithAuth,
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
