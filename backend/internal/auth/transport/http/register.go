package authhttp

import (
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req authdomain.RegisterRequest
	if err := helpers.ReadJSON(w, r, &req); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
		return
	}

	err := req.Validate()
	if err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
		return
	}

	ctx := r.Context()
	err = h.svc.Register(ctx, req)
	if err != nil {
		h.handleErrorRegisterResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "verify your email, before login"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, nil)
	}
}

func (h *Handler) handleErrorRegisterResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrUserAlreadyExists):
		if respErr := helpers.ErrorResponse(w, http.StatusConflict, "If this email is not yet registered, you will receive a confirmation link."); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	default:
		h.jsonLogger.Error(err, nil)
		if respErr := helpers.ServerErrorResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	}
}
