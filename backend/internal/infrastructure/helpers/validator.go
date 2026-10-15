package helpers

import (
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"
)

type validator interface {
	Validate() error
}

// DecodeAndValidate decodes req from the request body and validates it, writing an error
// response and returning false on failure.
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, jsonLogger *logger.Logger, req validator, beforeValidateHook func()) bool {
	if err := ReadJSON(w, r, req); err != nil {
		if respErr := BadRequestResponse(w, err); respErr != nil {
			jsonLogger.Error(respErr, map[string]any{
				"path":   r.URL.Path,
				"method": r.Method,
			})
		}
		return false
	}

	if beforeValidateHook != nil {
		beforeValidateHook()
	}

	if err := req.Validate(); err != nil {
		if respErr := BadRequestResponse(w, err); respErr != nil {
			jsonLogger.Error(respErr, map[string]any{
				"path":   r.URL.Path,
				"method": r.Method,
			})
		}
		return false
	}
	return true
}
