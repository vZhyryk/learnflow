package authhttp

import (
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		if respErr := helpers.ErrorResponse(w, http.StatusAccepted, "If this email is not yet registered, you will receive a confirmation link."); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "user_already_exists", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrAccountLocked):
		h.logAuthFailure(r, r.URL.Path, "account_locked", userIDProps(r))
		w.Header().Set("Retry-After", "900")
		if respErr := helpers.ErrorResponse(w, http.StatusTooManyRequests, "account temporarily locked"); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "account_locked", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrAccountBlocked):
		h.logAuthFailure(r, r.URL.Path, "account_blocked", userIDProps(r))
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "account is blocked", "code": "account_blocked"}); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "account_blocked", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrEmailNotVerified):
		h.logAuthFailure(r, r.URL.Path, "email_not_verified", userIDProps(r))
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "email not verified", "code": "email_not_verified"}); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "email_not_verified", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrTokenExpired):
		if respErr := helpers.BadRequestResponse(w, errors.New("token expired or invalid")); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "token_expired", "path": r.URL.Path})
		}

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
		if respErr := helpers.InvalidCredentialsResponse(w); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "invalid_credentials_group", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrWrongPassword):
		h.logAuthFailure(r, r.URL.Path, "wrong_password", userIDProps(r))
		if respErr := helpers.ErrorResponse(w, http.StatusUnprocessableEntity, "incorrect current password"); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "wrong_password", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrSamePassword):
		if respErr := helpers.ErrorResponse(w, http.StatusUnprocessableEntity, "new password must differ from current"); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "same_password", "path": r.URL.Path})
		}

	case errors.Is(err, authdomain.ErrInvalidAccountState):
		if respErr := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "if your account is eligible, you will receive an email"}, nil); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "invalid_account_state", "path": r.URL.Path})
		}

	default:
		h.jsonLogger.Error(err, map[string]any{
			"path":       r.URL.Path,
			"ip":         appcontext.IPAddressFromContext(r.Context()),
			"error_type": fmt.Sprintf("%T", err),
		})

		if respErr := helpers.ServerErrorResponse(w); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"case": "server_error_response_write", "path": r.URL.Path})
		}
	}
}
