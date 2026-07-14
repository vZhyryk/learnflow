package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req authdomain.RegisterRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	userID, err := h.svc.Register(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "verify your email, before login"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"event": registerEvent, "path": r.URL.Path})
	}

	h.logAuthEvent(r, registerEvent, map[string]any{"user_id": userID})
}
