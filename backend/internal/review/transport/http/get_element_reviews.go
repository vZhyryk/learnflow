package reviewhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
	"strconv"
)

func (h *Handler) listCourseReviews(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetCourseReviews(ctx, pagination.ParsePaginationParams(r), courseID, reviewdomain.ReviewFilter{})
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

	reviews, err := h.svc.GetContentReviews(ctx, pagination.ParsePaginationParams(r), contentID, reviewdomain.ReviewFilter{})
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) listCourseReviewsAdmin(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetCourseReviews(ctx, pagination.ParsePaginationParams(r), courseID, parseRatingFilter(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
func (h *Handler) listContentReviewsAdmin(w http.ResponseWriter, r *http.Request) {
	contentID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetContentReviews(ctx, pagination.ParsePaginationParams(r), contentID, parseRatingFilter(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) listArticleReviews(w http.ResponseWriter, r *http.Request) {
	articleID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetArticleReviews(ctx, pagination.ParsePaginationParams(r), articleID, reviewdomain.ReviewFilter{})
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) listArticleReviewsAdmin(w http.ResponseWriter, r *http.Request) {
	articleID := r.PathValue("id")
	ctx := r.Context()

	reviews, err := h.svc.GetArticleReviews(ctx, pagination.ParsePaginationParams(r), articleID, parseRatingFilter(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"reviews": reviews}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func parseRatingFilter(r *http.Request) reviewdomain.ReviewFilter {
	rating, err := strconv.Atoi(r.URL.Query().Get("rating"))
	if err != nil {
		rating = 0
	}

	return reviewdomain.ReviewFilter{
		Op:     r.URL.Query().Get("op"),
		Rating: rating,
	}
}
