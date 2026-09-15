package reviewhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	reviewdomain "learnflow_backend/internal/review/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/validator"
	"net/http"
)

func (h *Handler) deleteCourseReviewAdmin(w http.ResponseWriter, r *http.Request) {
	reviewID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(reviewID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidReviewID)
		return
	}

	err := h.svc.DeleteCourseReviewAdmin(ctx, reviewID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Course review deleted successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) deleteContentReviewAdmin(w http.ResponseWriter, r *http.Request) {
	reviewID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(reviewID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidReviewID)
		return
	}
	err := h.svc.DeleteContentReviewAdmin(ctx, reviewID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Content review deleted successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) deleteCourseReview(w http.ResponseWriter, r *http.Request) {
	reviewID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(reviewID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidReviewID)
		return
	}

	err := h.svc.DeleteCourseReview(ctx, reviewID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Course review deleted successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) deleteContentReview(w http.ResponseWriter, r *http.Request) {
	reviewID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(reviewID) {
		h.handleErrorResponse(w, r, reviewdomain.ErrInvalidReviewID)
		return
	}
	err := h.svc.DeleteContentReview(ctx, reviewID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Content review deleted successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
