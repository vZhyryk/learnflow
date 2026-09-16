package adminhttp

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) createAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req admindomain.CreateAnnouncementRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.CreatedByUserID = user.ID
	}) {
		return
	}

	announcementID, err := h.svc.CreateAnnouncement(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"announcement_id": announcementID}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req admindomain.UpdateAnnouncementRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UpdatedByUserID = user.ID
	}) {
		return
	}

	err := h.svc.UpdateAnnouncement(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "announcement was successfully updated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) approveAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	announcementID := r.PathValue("id")

	err := h.svc.ApproveAnnouncement(ctx, announcementID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "announcement was successfully approved"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) getAnnouncements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contentList, err := h.svc.GetAnnouncements(ctx, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"announcements": contentList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) getUnApprovedAnnouncements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contentList, err := h.svc.GetUnApprovedAnnouncements(ctx, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"announcements": contentList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) getApprovedAnnouncements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contentList, err := h.svc.GetApprovedAnnouncements(ctx, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"announcements": contentList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}

func (h *Handler) getExpiredAnnouncements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contentList, err := h.svc.GetExpiredAnnouncements(ctx, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"announcements": contentList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
