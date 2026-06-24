package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req authdomain.VerifyEmailRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()

	userID, err := h.svc.VerifyEmail(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "you can now login into platform"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"event": verifyEmailEvent, "user_id": userID, "path": r.URL.Path})
	}

	h.logAuthEvent(r, verifyEmailEvent, map[string]any{"user_id": userID})

}
