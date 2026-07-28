package contenthttp

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) listAllContentItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	status := contentdomain.ContentItemStatus(r.URL.Query().Get("status"))
	if status != "" && !status.Valid() {
		h.handleErrorResponse(w, r, contentdomain.ErrInvalidGetType)
		return
	}

	contentList, err := h.svc.GetAllContentItems(ctx, status, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"content_item_list": contentList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
