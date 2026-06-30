package authhttp

import (
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
	"time"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		// 202: prevents email enumeration; async outbox — email not sent synchronously.
		h.handleErrorRespond(r, "user_already_exists", func() error {
			return helpers.ErrorResponse(w, http.StatusAccepted, "If this email is not yet registered, you will receive a confirmation link.")
		})

	case errors.Is(err, authdomain.ErrAccountLocked):
		h.logAuthFailure(r, r.URL.Path, "account_locked", userIDProps(r))
		w.Header().Set("Retry-After", setAccountLockHeader(err))
		h.handleErrorRespond(r, "account_locked", func() error {
			return helpers.ErrorResponse(w, http.StatusTooManyRequests, "account temporarily locked")
		})
	case errors.Is(err, authdomain.ErrAccountBlocked):
		h.logAuthFailure(r, r.URL.Path, "account_blocked", userIDProps(r))
		h.handleErrorRespond(r, "account_blocked", func() error {
			return helpers.ForbiddenResponse(w, helpers.Envelope{"error": "account is blocked", "code": "account_blocked"})
		})
	case errors.Is(err, authdomain.ErrEmailNotVerified):
		h.logAuthFailure(r, r.URL.Path, "email_not_verified", userIDProps(r))
		h.handleErrorRespond(r, "email_not_verified", func() error {
			return helpers.ForbiddenResponse(w, helpers.Envelope{"error": "email not verified", "code": "email_not_verified"})
		})

	case errors.Is(err, authdomain.ErrTokenExpired):
		h.handleErrorRespond(r, "token_expired", func() error {
			return helpers.BadRequestResponse(w, errors.New("token expired or invalid"))
		})

	case errors.Is(err, authdomain.ErrUserNotFound),
		errors.Is(err, authdomain.ErrInvalidCredentials),
		errors.Is(err, authdomain.ErrEmailAlreadyInUse),
		errors.Is(err, authdomain.ErrTokenUsed),
		errors.Is(err, authdomain.ErrInvalidToken),
		errors.Is(err, authdomain.ErrInvalidCredentialFormat),
		errors.Is(err, authdomain.ErrSessionNotFound),
		errors.Is(err, authdomain.ErrSessionRevoked),
		errors.Is(err, authdomain.ErrSessionExpired):
		h.logAuthFailure(r, r.URL.Path, "invalid_credentials", userIDProps(r))
		h.handleErrorRespond(r, "invalid_credentials_group", func() error {
			return helpers.InvalidCredentialsResponse(w)
		})

	case errors.Is(err, authdomain.ErrWrongPassword):
		h.logAuthFailure(r, r.URL.Path, "wrong_password", userIDProps(r))
		h.handleErrorRespond(r, "wrong_password", func() error {
			return helpers.ErrorResponse(w, http.StatusUnprocessableEntity, "incorrect current password")
		})

	case errors.Is(err, authdomain.ErrSamePassword):
		h.handleErrorRespond(r, "same_password", func() error {
			return helpers.ErrorResponse(w, http.StatusUnprocessableEntity, "new password must differ from current")
		})

	case errors.Is(err, authdomain.ErrInvalidAccountState):
		h.handleErrorRespond(r, "invalid_account_state", func() error {
			return helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "if your account is eligible, you will receive an email"}, nil)
		})

	default:
		h.jsonLogger.Error(err, map[string]any{
			"path":       r.URL.Path,
			"ip":         appcontext.IPAddressFromContext(r.Context()),
			"error_type": fmt.Sprintf("%T", err),
		})
		h.handleErrorRespond(r, "server_error_response_write", func() error {
			return helpers.ServerErrorResponse(w)
		})
	}
}

func setAccountLockHeader(err error) string {
	var lockedErr *authdomain.ErrAccountLockedError
	if errors.As(err, &lockedErr) {
		if secs := int(time.Until(lockedErr.LockedUntil).Seconds()); secs > 0 {
			return fmt.Sprintf("%d", secs)
		}
	}
	return "900"
}

func (h *Handler) handleErrorRespond(r *http.Request, caseName string, fn func() error) {
	if err := fn(); err != nil {
		h.jsonLogger.Error(err, map[string]any{"case": caseName, "path": r.URL.Path})
	}
}
