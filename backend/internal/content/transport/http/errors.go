package contenthttp

import (
	"errors"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, contentdomain.ErrContentItemNotFound):
		h.handleErrorRespond(r, "content_item_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, contentdomain.ErrInvalidSlug),
		errors.Is(err, contentdomain.ErrInvalidTitle),
		errors.Is(err, contentdomain.ErrInvalidDescription),
		errors.Is(err, contentdomain.ErrInvalidThumbnailURL),
		errors.Is(err, contentdomain.ErrInvalidEstimatedMinutes),
		errors.Is(err, contentdomain.ErrInvalidSeoTitle),
		errors.Is(err, contentdomain.ErrInvalidSeoDescription),
		errors.Is(err, contentdomain.ErrInvalidOgImageURL),
		errors.Is(err, contentdomain.ErrInvalidCanonicalURL),
		errors.Is(err, contentdomain.ErrInvalidContentItemID),
		errors.Is(err, contentdomain.ErrInvalidContentItemStatus),
		errors.Is(err, contentdomain.ErrInvalidGetType),
		errors.Is(err, contentdomain.ErrInvalidEstimatedPages),
		errors.Is(err, contentdomain.ErrInvalidMedia):

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
