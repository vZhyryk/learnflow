package articlehttp

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) listArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articleList, err := h.svc.GetAllArticles(ctx, articledomain.PublishedStatus, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"article_list": articleList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
