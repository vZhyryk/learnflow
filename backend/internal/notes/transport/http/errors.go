package noteshttp

import (
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/helpers"
	notesdomain "learnflow_backend/internal/notes/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, notesdomain.ErrResourceDataMisMatch),
		errors.Is(err, notesdomain.ErrInvalidTitle),
		errors.Is(err, notesdomain.ErrInvalidDescription),
		errors.Is(err, notesdomain.ErrInvalidBody),
		errors.Is(err, notesdomain.ErrInvalidResourceType),
		errors.Is(err, notesdomain.ErrInvalidResourceID),
		errors.Is(err, notesdomain.ErrInvalidUserID),
		errors.Is(err, notesdomain.ErrInvalidSearch):
		h.handleErrorRespond(r, "validation_error", func() error {
			return helpers.ErrorResponse(w, http.StatusUnprocessableEntity, helpers.RootError(err).Error())
		})
	case errors.Is(err, notesdomain.ErrNoteNotFound):
		h.handleErrorRespond(r, "note_not_found", func() error {
			return helpers.NotFoundResponse(w)
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

// handleErrorRespond runs fn and logs (never returns) a failure to write the response —
// by this point the handler has nothing left to do about it.
func (h *Handler) handleErrorRespond(r *http.Request, caseName string, fn func() error) {
	helpers.LogRespondError(h.jsonLogger, r, caseName, nil, fn)
}
