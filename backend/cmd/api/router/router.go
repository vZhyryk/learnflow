// Package router wires all HTTP routes to their handlers.
package router

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/internal/access"
	"learnflow_backend/internal/admin"
	adminrepository "learnflow_backend/internal/admin/repository"
	adminservice "learnflow_backend/internal/admin/service"
	"learnflow_backend/internal/article"
	articlerepository "learnflow_backend/internal/article/repository"
	articleservice "learnflow_backend/internal/article/service"
	auditrepository "learnflow_backend/internal/audit/repository"
	"learnflow_backend/internal/auth"
	authdomain "learnflow_backend/internal/auth/domain"
	authrepository "learnflow_backend/internal/auth/repository"
	authservice "learnflow_backend/internal/auth/service"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/content"
	contentrepository "learnflow_backend/internal/content/repository"
	contentservice "learnflow_backend/internal/content/service"
	"learnflow_backend/internal/courses"
	courserepository "learnflow_backend/internal/courses/repository"
	courseservice "learnflow_backend/internal/courses/service"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/helpers"
	"learnflow_backend/internal/notes"
	notesrepository "learnflow_backend/internal/notes/repository"
	notesservice "learnflow_backend/internal/notes/service"
	"learnflow_backend/internal/review"
	reviewrepository "learnflow_backend/internal/review/repository"
	reviewservice "learnflow_backend/internal/review/service"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/tokens"
	"learnflow_backend/internal/users"
	usersrepository "learnflow_backend/internal/users/repository"
	usersservice "learnflow_backend/internal/users/service"
	"net/http"
	"strings"
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

	adminAction := auditrepository.New(a.DB)

	transactor := db.NewTransactor(a.DB)
	outbox := events.NewOutboxWriter(a.DB)

	chains := route.buildChains()
	adminStaticWithAuth := chains.StaticWithAuth.Append(route.RequireRole(authdomain.RoleAdmin, authdomain.RoleSubAdmin))

	// Auth Routes
	authRepo := authrepository.NewRepository(a.DB)

	authSvc, err := authservice.New(
		authservice.Repos{
			UserRepo:    authRepo,
			SessionRepo: authRepo,
			TokenRepo:   authRepo,
			Transactor:  transactor,
		},
		authservice.Utils{
			Outbox:    outbox,
			Token:     route.token,
			Blocklist: a.Redis,
			Audit:     adminAction,
		},
		authservice.Options{})
	if err != nil {
		return nil, fmt.Errorf("router: NewRouter: %w", err)
	}
	auth.RegisterAuthRoutes(router, authSvc, chains, a.Logger)

	// User Routes
	usersrepo := usersrepository.NewRepository(a.DB)
	userSvc := usersservice.New(usersrepo)
	users.RegisterUsersRoutes(router, userSvc, chains.StaticWithAuth, a.Logger)

	// Course Routes
	courseRepo := courserepository.NewRepository(a.DB)
	courseSvc := courseservice.New(courseRepo, adminAction, transactor)
	courses.RegisterCourseRoutes(router, courseSvc, chains.Static, adminStaticWithAuth, a.Logger)

	// Content Routes
	contentRepo := contentrepository.NewRepository(a.DB)
	contentSvc := contentservice.New(contentRepo, adminAction, transactor)
	content.RegisterContentRoutes(router, contentSvc, chains.Static, adminStaticWithAuth, a.Logger)

	// Article Routes
	articleRepo := articlerepository.NewRepository(a.DB)
	articleSvc := articleservice.New(articleRepo, adminAction, transactor)
	article.RegisterArticleRoutes(router, articleSvc, chains.Static, adminStaticWithAuth, a.Logger)

	// Review Routes
	reviewRepo := reviewrepository.NewRepository(a.DB)
	accessChecker := access.New(a.DB)
	reviewSvc := reviewservice.New(reviewRepo, reviewRepo, reviewRepo, transactor, accessChecker, adminAction)
	review.RegisterReviewRoutes(router, reviewSvc, chains.Static, chains.StaticWithAuth, adminStaticWithAuth, a.Logger)

	// Notes Routes
	notesRepo := notesrepository.NewRepository(a.DB)
	notesSvc := notesservice.New(notesRepo, courseRepo, contentRepo, transactor)
	notes.RegisterNotesRoutes(router, notesSvc, chains.StaticWithAuth, a.Logger)

	// Admin Routes
	adminRepo := adminrepository.NewRepository(a.DB)
	adminSvc := adminservice.New(
		adminservice.Repos{
			AnnounRepo:      adminRepo,
			UserRepo:        adminRepo,
			ActionRepo:      adminAction,
			SessionRepo:     authRepo,
			CourseRepo:      courseRepo,
			ContentItemRepo: contentRepo,
		}, adminservice.Utils{
			Transactor: transactor,
			Outbox:     outbox,
			Blocklist:  a.Redis,
		})
	admin.RegisterAdminRoutes(router, adminSvc, adminStaticWithAuth, chains.StaticWithAuth, a.Logger)

	route.registerHelperRoutes(router)

	return route, nil
}

// registerHelperRoutes registers the JSON 404 fallback and the unauthenticated health, readiness and metrics endpoints.
func (route *RouteHandler) registerHelperRoutes(router *http.ServeMux) {
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helpers.LogRespondError(route.App.Logger, r, "not_found_response_write", nil, func() error {
			return helpers.NotFoundResponse(w)
		})
	}))

	router.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helpers.LogRespondError(route.App.Logger, r, "health_response_write", map[string]any{"method": r.Method}, func() error {
			return helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"status": "ok"}, nil)
		})
	}))

	router.Handle("GET /readiness", http.HandlerFunc(route.Readiness))

	router.Handle("GET /metrics", promhttp.Handler())
}

const rateLimitBodyLimit = 4_096

// bodyRateLimitKey reads a bounded body via extractField to key the limiter, then resets
// r.Body for downstream handlers. Do NOT add body-reading middleware after this call.
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

	return route.rateLimitDigest(value)
}

// rateLimitDigest keys a limiter bucket by an HMAC of value, so a leaked key cannot be reversed by a dictionary of emails.
func (route *RouteHandler) rateLimitDigest(value string) string {
	mac := hmac.New(sha256.New, []byte(route.App.Config.Secret.JWTSecret))
	mac.Write([]byte("ratelimit-key:" + value))

	return hex.EncodeToString(mac.Sum(nil))
}

func (route *RouteHandler) getEmailFromBody(r *http.Request) string {
	return route.bodyRateLimitKey(r, func(bodyBytes []byte) (string, bool) {
		var req authdomain.RequestPasswordResetRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			return "", false
		}
		return strings.ToLower(strings.TrimSpace(req.Email)), true
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

func ipKey(r *http.Request) string {
	return appcontext.IPAddressFromContext(r.Context())
}

func (route *RouteHandler) ipEmailKey(r *http.Request) string {
	return ipKey(r) + ":" + route.getEmailFromBody(r)
}

func (route *RouteHandler) ipTokenKey(r *http.Request) string {
	return ipKey(r) + ":" + route.getTokenFromBody(r)
}

// buildChains sets up the auth route limits; login and password reset also have a looser per-IP limit,
// so one IP cannot try endless different emails.
func (route *RouteHandler) buildChains() authhttp.AuthRouteChains {
	cfg := route.App.Config.Limiter

	static := route.NewRouteRateLimiter("static", cfg.Rps, time.Second, cfg.Burst, ipKey)
	register := route.NewRouteRateLimiter("register", 3, time.Hour, 3, ipKey)
	emailVerify := route.NewRouteRateLimiter("email_verify", 3, time.Hour, 3, route.ipTokenKey)

	loginPerIP := route.NewRouteRateLimiter("login_ip", 20, time.Minute, 20, ipKey)
	loginPerEmail := route.NewRouteRateLimiter("login", 5, time.Minute, 5, route.ipEmailKey)

	resetPerIP := route.NewRouteRateLimiter("password_reset_ip", 10, time.Hour, 10, ipKey)
	resetPerEmail := route.NewRouteRateLimiter("password_reset", 2, time.Hour, 2, route.ipEmailKey)

	return authhttp.AuthRouteChains{
		Static:         route.SetChain(static),
		Login:          route.SetChain(loginPerIP, loginPerEmail),
		Register:       route.SetChain(register),
		PassReset:      route.SetChain(resetPerIP, resetPerEmail),
		EmailVerify:    route.SetChain(emailVerify),
		StaticWithAuth: route.SetChain(static).Append(route.AuthenticateUser),
	}
}

// Readiness checks DB and Redis connectivity and returns 200 if the service is ready to handle traffic.
func (h *RouteHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.App.DB.Ping(ctx); err != nil {
		if respErr := helpers.WriteJSON(w, http.StatusServiceUnavailable, helpers.Envelope{"status": "unavailable", "reason": "database"}, nil); respErr != nil {
			h.App.Logger.Error(respErr, map[string]any{
				"status":   http.StatusServiceUnavailable,
				"envelope": helpers.Envelope{"status": "unavailable", "reason": "database"},
			})
		}
		return
	}
	if err := h.App.Redis.Raw().Ping(ctx).Err(); err != nil {
		if respErr := helpers.WriteJSON(w, http.StatusServiceUnavailable, helpers.Envelope{"status": "unavailable", "reason": "redis"}, nil); respErr != nil {
			h.App.Logger.Error(respErr, map[string]any{
				"status":   http.StatusServiceUnavailable,
				"envelope": helpers.Envelope{"status": "unavailable", "reason": "redis"},
			})
		}
		return
	}
	if respErr := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"status": "ready"}, nil); respErr != nil {
		h.App.Logger.Error(respErr, map[string]any{
			"status":   http.StatusOK,
			"envelope": helpers.Envelope{"status": "ready"},
		})
	}
}

// SetChain builds the standard middleware chain, inserting the limiters (in order) for rate limiting.
func (route *RouteHandler) SetChain(limiters ...alice.Constructor) alice.Chain {
	return alice.New(route.RecoverPanic, route.SetIPAddress, route.SetRequestID, route.RequestLogger).
		Append(limiters...).
		Append(route.EnableCORS, route.Timeout, route.SetSecurityHeaders)
}
