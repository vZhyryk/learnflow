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
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	reviewID, err := h.svc.CreateCourseReviewAdmin(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"review_id": reviewID}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createContentReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateContentReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	reviewID, err := h.svc.CreateContentReviewAdmin(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"review_id": reviewID}, nil)
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

	reviewID, err := h.svc.CreateCourseReview(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"review_id": reviewID}, nil)
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

	reviewID, err := h.svc.CreateContentReview(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"review_id": reviewID}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) createArticleReviewAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req reviewdomain.CreateArticleReviewRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	reviewID, err := h.svc.CreateArticleReviewAdmin(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"review_id": reviewID}, nil)
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

	reviewID, err := h.svc.CreateArticleReview(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"review_id": reviewID}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
