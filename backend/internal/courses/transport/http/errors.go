package coursehttp

import (
	"errors"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, coursedomain.ErrCourseNotFound):
		h.handleErrorRespond(r, "course_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, coursedomain.ErrInvalidSlug),
		errors.Is(err, coursedomain.ErrInvalidTitle),
		errors.Is(err, coursedomain.ErrInvalidDescription),
		errors.Is(err, coursedomain.ErrInvalidThumbnailURL),
		errors.Is(err, coursedomain.ErrInvalidPreviewVideoURL),
		errors.Is(err, coursedomain.ErrInvalidEstimatedMinutes),
		errors.Is(err, coursedomain.ErrInvalidSeoTitle),
		errors.Is(err, coursedomain.ErrInvalidSeoDescription),
		errors.Is(err, coursedomain.ErrInvalidOgImageURL),
		errors.Is(err, coursedomain.ErrInvalidCanonicalURL),
		errors.Is(err, coursedomain.ErrInvalidCourseID),
		errors.Is(err, coursedomain.ErrInvalidCourseStatus),
		errors.Is(err, coursedomain.ErrInvalidGetType):
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
