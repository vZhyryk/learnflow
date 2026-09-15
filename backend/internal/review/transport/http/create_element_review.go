package reviewhttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	reviewdomain "learnflow_backend/internal/review/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) createCourseReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateCourseReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	if err := h.svc.CreateCourseReviewAdmin(ctx, req); err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Course review created successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createContentReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateContentReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	if err := h.svc.CreateContentReviewAdmin(ctx, req); err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Content review created successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createCourseReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateCourseReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	if err := h.svc.CreateCourseReview(ctx, req); err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Course review created successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createContentReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateContentReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	if err := h.svc.CreateContentReview(ctx, req); err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Content review created successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createArticleReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateArticleReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	if err := h.svc.CreateArticleReviewAdmin(ctx, req); err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Article review created successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createArticleReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateArticleReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	if err := h.svc.CreateArticleReview(ctx, req); err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Article review created successfully"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
