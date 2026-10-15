package contenthttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) getContentItemBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	ctx := r.Context()
	contentItem, err := h.svc.GetContentItemBySlug(ctx, slug)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"content_item": contentItem}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
