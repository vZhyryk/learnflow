package authhttp

import (
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req authdomain.LoginRequest
	if err := helpers.ReadJSON(w, r, &req); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
		return
	}
	ctx := r.Context()
	ua := r.UserAgent()
	req.IPAddress = appcontext.IPAddressFromContext(r.Context())
	req.UserAgent = ua

	err := req.Validate()
	if err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
		return
	}

	tokens, err := h.svc.Login(ctx, req)
	if err != nil {
		h.handleErrorLoginResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"tokens": tokens}, nil)
	if err != nil {
		h.jsonLogger.Error(err, nil)
	}
}

func (h *Handler) handleErrorLoginResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		if respErr := helpers.InvalidCredentialsResponse(w); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	case errors.Is(err, authdomain.ErrAccountLocked):
		w.Header().Set("Retry-After", "900")
		if respErr := helpers.ErrorResponse(w, http.StatusTooManyRequests, "account temporarily locked"); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	case errors.Is(err, authdomain.ErrAccountBlocked):
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "account is blocked", "code": "account_blocked"}); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	case errors.Is(err, authdomain.ErrEmailNotVerified):
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "email not verified", "code": "email_not_verified"}); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	default:
		h.jsonLogger.Error(err, nil)
		if respErr := helpers.ServerErrorResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	}
}
