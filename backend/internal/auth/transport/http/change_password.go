package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var req authdomain.ChangePasswordRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	user, ok := appcontext.UserFromContext(ctx)
	if !ok {
		h.handleErrorResponse(w, r, authdomain.ErrUserNotFound)
		return
	}
	req.UserID = user.ID

	req.JTI = appcontext.JTIFromContext(ctx)
	req.AccessTokenExpiresAt = appcontext.AccessTokenExpiresAtFromContext(ctx)

	err := h.svc.ChangePassword(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "password has been changed"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
	h.logAuthEvent(r, changePasswordEvent, map[string]any{"user_id": user.ID})
}
