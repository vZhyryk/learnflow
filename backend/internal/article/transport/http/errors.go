package articlehttp

import (
	"errors"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) handleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, articledomain.ErrArticleNotFound):
		h.handleErrorRespond(r, "article_not_found", func() error {
			return helpers.NotFoundResponse(w)
		})

	case errors.Is(err, articledomain.ErrInvalidSlug),
		errors.Is(err, articledomain.ErrInvalidTitle),
		errors.Is(err, articledomain.ErrInvalidDescription),
		errors.Is(err, articledomain.ErrInvalidSeoTitle),
		errors.Is(err, articledomain.ErrInvalidSeoDescription),
		errors.Is(err, articledomain.ErrInvalidOgImageURL),
		errors.Is(err, articledomain.ErrInvalidArticleID),
		errors.Is(err, articledomain.ErrInvalidArticleStatus),
		errors.Is(err, articledomain.ErrInvalidGetType):
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
