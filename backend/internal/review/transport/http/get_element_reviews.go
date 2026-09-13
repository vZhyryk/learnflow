package reviewhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) listCourseReviews(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetCourseReviews(ctx, pagination.ParsePaginationParams(r), courseID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
func (h *Handler) listContentReviews(w http.ResponseWriter, r *http.Request) {
	contentID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetContentReviews(ctx, pagination.ParsePaginationParams(r), contentID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
