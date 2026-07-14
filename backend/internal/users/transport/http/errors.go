package usershttp

import (
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	usersdomain "learnflow_backend/internal/users/domain"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, usersdomain.ErrUserNotFound):
		h.handleErrorRespond(r, "user_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, usersdomain.ErrFirstNameInvalid),
		errors.Is(err, usersdomain.ErrLastNameTooLong),
		errors.Is(err, usersdomain.ErrPhoneNumberInvalid),
		errors.Is(err, usersdomain.ErrCountryInvalid),
		errors.Is(err, usersdomain.ErrBioTooLong),
		errors.Is(err, usersdomain.ErrAvatarURLInvalid),
		errors.Is(err, usersdomain.ErrGenderInvalid),
		errors.Is(err, usersdomain.ErrUILanguageInvalid),
		errors.Is(err, usersdomain.ErrTimezoneInvalid),
		errors.Is(err, usersdomain.ErrDateOfBirthInvalid):
		h.handleErrorRespond(r, "validation_error", func() error {
			return helpers.ErrorResponse(w, http.StatusUnprocessableEntity, err.Error())
		})

	default:
		h.jsonLogger.Error(err, map[string]any{
			"path":       r.URL.Path,
			"ip":         appcontext.IPAddressFromContext(r.Context()),
			"error_type": fmt.Sprintf("%T", err),
		})
		h.handleErrorRespond(r, "server_error_response_write", func() error {
			return helpers.ServerErrorResponse(w)
		})
	}
}

// handleErrorRespond runs fn (a response-writing call) and logs a failure to write the
// response itself. The write error is only logged, never returned — by this point the
// handler has already decided what to respond with, and callers have nothing left to do.
func (h *Handler) handleErrorRespond(r *http.Request, caseName string, fn func() error) {
	helpers.LogRespondError(h.jsonLogger, r, caseName, nil, fn)
}
