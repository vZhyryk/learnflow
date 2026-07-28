package contenthttp

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) updateContentItem(w http.ResponseWriter, r *http.Request) {
	var req contentdomain.UpdateContentItemRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	err := h.svc.UpdateContentItem(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "content item was successfully updated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
