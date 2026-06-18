package authhttp

import (
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ctx := r.Context()

	req := authdomain.VerifyEmailRequest{
		Token: token,
	}

	err := req.Validate()
	if err != nil {
		h.handleErrorEmailVerifyResponse(w, err)
		return
	}

	err = h.svc.VerifyEmail(ctx, req)
	if err != nil {
		h.handleErrorEmailVerifyResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "you can now login into platform"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, nil)
	}
}

func (h *Handler) handleErrorEmailVerifyResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrTokenUsed), errors.Is(err, authdomain.ErrInvalidCredentialFormat), errors.Is(err, authdomain.ErrInvalidToken):
		if respErr := helpers.InvalidCredentialsResponse(w); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}

	case errors.Is(err, authdomain.ErrTokenExpired):
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}

	default:
		h.jsonLogger.Error(err, nil)
		if respErr := helpers.ServerErrorResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	}
}
