package articlehttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) getArticleBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	ctx := r.Context()
	article, err := h.svc.GetArticleBySlug(ctx, slug)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"article": article}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
