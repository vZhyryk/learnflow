package helpers

import (
	"net/http"
)

// InvalidCredentialsResponse writes a 401 Unauthorized response.
func InvalidCredentialsResponse(w http.ResponseWriter) error {
	return WriteJSON(w, http.StatusUnauthorized, Envelope{"error": "unauthorized", "code": "unauthorized"}, nil)
}

// ForbiddenResponse writes a 403 Forbidden response with the given envelope message.
func ForbiddenResponse(w http.ResponseWriter, msg Envelope) error {
	return WriteJSON(w, http.StatusForbidden, msg, nil)
}

// ServerErrorResponse writes a 500 Internal Server Error response with the given error.
func ServerErrorResponse(w http.ResponseWriter) error {
	return WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "internal server error"}, nil)
}

// BadRequestResponse writes a 400 Bad Request response with the given error.
func BadRequestResponse(w http.ResponseWriter, err error) error {
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, Envelope{"error": err.Error()}, nil)
	}
	return WriteJSON(w, http.StatusBadRequest, Envelope{"error": "bad request"}, nil)
}

// ErrorResponse writes a response with the given status code and error message.
func ErrorResponse(w http.ResponseWriter, status int, msg string) error {
	return WriteJSON(w, status, Envelope{"error": msg}, nil)
}

// NotFoundResponse writes a 404 Not Found response.
func NotFoundResponse(w http.ResponseWriter) error {
	return ErrorResponse(w, http.StatusNotFound, "the requested resource could not be found")
}
