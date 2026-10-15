package reviewhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/validator"
	"net/http"
)

func (h *Handler) courseReviewStats(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	ctx := r.Context()

	if !validator.IsValidUUID(courseID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidCourseID)
		return
	}

	rating, count, err := h.svc.GetCourseReviewStats(ctx, courseID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"rating": rating, "count": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) contentReviewStats(w http.ResponseWriter, r *http.Request) {
	contentID := r.PathValue("id")
	ctx := r.Context()

	if !validator.IsValidUUID(contentID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidContentItemID)
		return
	}

	rating, count, err := h.svc.GetContentReviewStats(ctx, contentID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"rating": rating, "count": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) articleReviewStats(w http.ResponseWriter, r *http.Request) {
	articleID := r.PathValue("id")
	ctx := r.Context()

	if !validator.IsValidUUID(articleID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidArticleID)
		return
	}

	rating, count, err := h.svc.GetArticleReviewStats(ctx, articleID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"rating": rating, "count": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
