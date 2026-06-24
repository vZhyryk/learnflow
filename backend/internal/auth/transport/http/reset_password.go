package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) initiatePasswordReset(w http.ResponseWriter, r *http.Request) {
	var req authdomain.RequestPasswordResetRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	err := h.svc.InitiatePasswordReset(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "password reset has been initiated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req authdomain.ResetPasswordRequest
	if err := helpers.ReadJSON(w, r, &req); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path, "method": r.Method})
		}
		return
	}

	err := req.Validate()
	if err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"path": r.URL.Path, "method": r.Method})
		}
		return
	}

	ctx := r.Context()
	err = h.svc.ResetPassword(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "password has been changed"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
