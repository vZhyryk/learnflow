package adminhttp

import (
	"errors"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, admindomain.ErrAnnouncementNotFound):
		h.handleErrorRespond(r, "announcement_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, admindomain.ErrForbiddenUserAction):
		h.handleErrorRespond(r, "forbidden_user_action", func() error {
			return helpers.ForbiddenResponse(w, helpers.Envelope{"error": err.Error(), "code": "forbidden_user_action"})
		})

	case errors.Is(err, admindomain.ErrInvalidUserState):
		h.handleErrorRespond(r, "invalid_user_state", func() error {
			return helpers.ErrorResponse(w, http.StatusConflict, err.Error())
		})

	case errors.Is(err, admindomain.ErrUserNotFound):
		h.handleErrorRespond(r, "user_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, admindomain.ErrInvalidID),
		errors.Is(err, admindomain.ErrInvalidGetType),
		errors.Is(err, admindomain.ErrInvalidBody),
		errors.Is(err, admindomain.ErrInvalidTitle),
		errors.Is(err, admindomain.ErrInvalidEntityID),
		errors.Is(err, admindomain.ErrInvalidChannel),
		errors.Is(err, admindomain.ErrInvalidEntityType),
		errors.Is(err, admindomain.ErrInvalidExpiresAt),
		errors.Is(err, admindomain.ErrEntityDataMisMatch),
		errors.Is(err, admindomain.ErrAnnouncementApproved):

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

// handleErrorRespond runs fn and logs (never returns) a failure to write the response —
// by this point the handler has nothing left to do about it.
func (h *Handler) handleErrorRespond(r *http.Request, caseName string, fn func() error) {
	helpers.LogRespondError(h.jsonLogger, r, caseName, nil, fn)
}
