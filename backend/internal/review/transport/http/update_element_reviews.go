package reviewhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	reviewdomain "learnflow_backend/internal/review/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) updateCourseReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req reviewdomain.UpdateCourseReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	err := h.svc.UpdateCourseReviewAdmin(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Course review updated successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateContentReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req reviewdomain.UpdateContentReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	err := h.svc.UpdateContentReviewAdmin(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Content review updated successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateCourseReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req reviewdomain.UpdateCourseReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	err := h.svc.UpdateCourseReview(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Course review updated successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateContentReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req reviewdomain.UpdateContentReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	err := h.svc.UpdateContentReview(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Content review updated successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateArticleReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req reviewdomain.UpdateArticleReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	err := h.svc.UpdateArticleReviewAdmin(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Article review updated successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateArticleReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req reviewdomain.UpdateArticleReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	err := h.svc.UpdateArticleReview(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Article review updated successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
