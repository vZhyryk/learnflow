package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) initRecoverAccount(w http.ResponseWriter, r *http.Request) {
	var req authdomain.RequestRecoverAccountRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	err := h.svc.InitRecoverAccount(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "account recover process was initiated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
	h.logAuthEvent(r, initRecoverAccountEvent, nil)
}

func (h *Handler) recoverAccount(w http.ResponseWriter, r *http.Request) {
	var req authdomain.RecoverAccountRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	err := h.svc.RecoverAccount(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "account has been recovered"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
	h.logAuthEvent(r, recoverAccountEvent, nil)
}
