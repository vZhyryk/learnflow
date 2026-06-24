package authhttp

import (
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ua := r.UserAgent()

	var req authdomain.LoginRequest
	if !h.decodeAndValidate(w, r, &req, func() {
		req.IPAddress = appcontext.IPAddressFromContext(r.Context())
		req.UserAgent = ua
	}) {
		return
	}

	tokens, err := h.svc.Login(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, authdomain.ErrInvalidCredentials):
			h.logAuthFailure(r, loginEvent, "invalid_credentials", map[string]any{})
		case errors.Is(err, authdomain.ErrAccountLocked):
			h.logAuthFailure(r, loginEvent, "account_locked", map[string]any{})
		case errors.Is(err, authdomain.ErrAccountBlocked):
			h.logAuthFailure(r, loginEvent, "account_blocked", map[string]any{})
		}
		h.handleErrorResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"auth": tokens}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"event": loginEvent, "path": r.URL.Path})
	}

	h.logAuthEvent(r, loginEvent, map[string]any{"user_id": tokens.UserID})
}
