package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var req authdomain.LogoutRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}
	ctx := r.Context()

	req.JTI = appcontext.JTIFromContext(ctx)
	req.AccessTokenExpiresAt = appcontext.AccessTokenExpiresAtFromContext(ctx)

	userID, err := h.svc.Logout(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "you have been logged out"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"event": logoutEvent, "user_id": userID, "path": r.URL.Path})
	}

	h.logAuthEvent(r, logoutEvent, map[string]any{"user_id": userID})
}
