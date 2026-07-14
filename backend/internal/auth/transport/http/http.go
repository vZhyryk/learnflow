package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"

	"github.com/justinas/alice"
)

const (
	loginEvent               = "auth.login"
	logoutEvent              = "auth.logout"
	refreshEvent             = "auth.refresh"
	registerEvent            = "auth.register"
	verifyEmailEvent         = "auth.verify_email"
	changePasswordEvent      = "auth.change_password"
	initiateEmailChangeEvent = "auth.initiate_email_change"
	changeEmailEvent         = "auth.change_email"
	initiatePassResetEvent   = "auth.initiate_pass_reset"
	resetPasswordEvent       = "auth.reset_password"
	initRecoverAccountEvent  = "auth.init_recover_account"
	recoverAccountEvent      = "auth.recover_account"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	svc        authdomain.Service
	jsonLogger *logger.Logger
}

// AuthRouteChains holds pre-configured middleware chains for different auth routes.
type AuthRouteChains struct {
	Static         alice.Chain
	Login          alice.Chain
	Register       alice.Chain
	PassReset      alice.Chain
	EmailVerify    alice.Chain
	StaticWithAuth alice.Chain
}

// NewHTTPHandler returns a new auth HTTP Handler.
func NewHTTPHandler(svc authdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes mounts all auth endpoints onto the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, chains AuthRouteChains) {
	mux.Handle("POST /api/v1/auth/login", chains.Login.ThenFunc(h.login))

	mux.Handle("POST /api/v1/auth/register", chains.Register.ThenFunc(h.register))

	mux.Handle("POST /api/v1/users/auth/email/verify", chains.EmailVerify.ThenFunc(h.verifyEmail))

	mux.Handle("POST /api/v1/auth/logout", chains.StaticWithAuth.ThenFunc(h.logout))

	mux.Handle("POST /api/v1/auth/refresh", chains.Static.ThenFunc(h.refresh))

	mux.Handle("POST /api/v1/users/auth/password/reset", chains.PassReset.ThenFunc(h.initiatePasswordReset))

	mux.Handle("PUT /api/v1/users/auth/password/reset", chains.PassReset.ThenFunc(h.resetPassword))

	mux.Handle("PUT /api/v1/users/auth/password/change", chains.StaticWithAuth.ThenFunc(h.changePassword))

	mux.Handle("POST /api/v1/users/auth/email/change", chains.StaticWithAuth.ThenFunc(h.initiateEmailChange))

	mux.Handle("PUT /api/v1/users/auth/email/change", chains.StaticWithAuth.ThenFunc(h.changeEmail))

	mux.Handle("POST /api/v1/users/auth/account/recover", chains.PassReset.ThenFunc(h.initRecoverAccount))

	mux.Handle("PUT /api/v1/users/auth/account/recover", chains.PassReset.ThenFunc(h.recoverAccount))
}
