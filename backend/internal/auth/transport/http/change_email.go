package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) initiateEmailChange(w http.ResponseWriter, r *http.Request) {
	var req authdomain.RequestEmailChangeRequest
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

	err := h.svc.InitiateEmailChange(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "email change process was initiated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
	h.logAuthEvent(r, initiateEmailChangeEvent, map[string]any{"user_id": user.ID})
}

func (h *Handler) changeEmail(w http.ResponseWriter, r *http.Request) {
	var req authdomain.EmailChangeRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	user, ok := appcontext.UserFromContext(ctx)
	if !ok {
		h.handleErrorResponse(w, r, authdomain.ErrInvalidCredentials)
		return
	}

	req.UserID = user.ID
	req.JTI = appcontext.JTIFromContext(ctx)
	req.AccessTokenExpiresAt = appcontext.AccessTokenExpiresAtFromContext(ctx)

	err := h.svc.ChangeEmail(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "email has been changed"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
	h.logAuthEvent(r, changeEmailEvent, map[string]any{"user_id": user.ID})
}
