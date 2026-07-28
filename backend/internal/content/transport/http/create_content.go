package contenthttp

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) createContentItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req contentdomain.CreateContentItemRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.CreatedByUserID = user.ID
	}) {
		return
	}

	contentItemID, err := h.svc.CreateContentItem(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"content_item_id": contentItemID}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
