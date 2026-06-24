package authhttp

import (
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		if respErr := helpers.ErrorResponse(w, http.StatusConflict, "If this email is not yet registered, you will receive a confirmation link."); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrAccountLocked):
		w.Header().Set("Retry-After", "900")
		if respErr := helpers.ErrorResponse(w, http.StatusTooManyRequests, "account temporarily locked"); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrAccountBlocked):
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "account is blocked", "code": "account_blocked"}); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrEmailNotVerified):
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "email not verified", "code": "email_not_verified"}); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrTokenExpired):
		if respErr := helpers.BadRequestResponse(w, errors.New("token expired or invalid")); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
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
		if respErr := helpers.InvalidCredentialsResponse(w); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrWrongPassword):
		if respErr := helpers.ErrorResponse(w, http.StatusUnprocessableEntity, "incorrect current password"); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrSamePassword):
		if respErr := helpers.ErrorResponse(w, http.StatusUnprocessableEntity, "new password must differ from current"); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	case errors.Is(err, authdomain.ErrInvalidAccountState):
		if respErr := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "if your account is eligible, you will receive an email"}, nil); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}

	default:
		h.jsonLogger.Error(err, nil)
		if respErr := helpers.ServerErrorResponse(w); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}
	}
}
