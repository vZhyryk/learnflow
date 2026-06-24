package authhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

type validator interface {
	Validate() error
}

func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, req validator, beforeValidateHook func()) bool {
	if err := helpers.ReadJSON(w, r, req); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}
		return false
	}

	if beforeValidateHook != nil {
		beforeValidateHook()
	}

	if err := req.Validate(); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(respErr, nil)
		}
		return false
	}
	return true
}
