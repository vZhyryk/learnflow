package consultationshttp

import (
	"errors"
	"fmt"
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, consultationsdomain.ErrSlotNotFound):
		h.handleErrorRespond(r, "slot_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, consultationsdomain.ErrSlotNotAvailable):
		h.handleErrorRespond(r, "slot_not_available", func() error {
			return helpers.ErrorResponse(w, http.StatusConflict, consultationsdomain.ErrSlotNotAvailable.Error())
		})

	case errors.Is(err, consultationsdomain.ErrInvalidSlotStatus),
		errors.Is(err, consultationsdomain.ErrInvalidSlotID),
		errors.Is(err, consultationsdomain.ErrDuplicateSlotID),
		errors.Is(err, consultationsdomain.ErrEmptySlotUpdate),
		errors.Is(err, consultationsdomain.ErrTooManySlots),
		errors.Is(err, consultationsdomain.ErrInvalidConsultant),
		errors.Is(err, consultationsdomain.ErrInvalidDateRange),
		errors.Is(err, consultationsdomain.ErrDateRangeTooLong):

		h.handleErrorRespond(r, "validation_error", func() error {
			return helpers.ErrorResponse(w, http.StatusUnprocessableEntity, helpers.RootError(err).Error())
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
