package contenthttp

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) listContentItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contentList, err := h.svc.GetAllContentItems(ctx, contentdomain.PublishedStatus, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"content_item_list": contentList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
