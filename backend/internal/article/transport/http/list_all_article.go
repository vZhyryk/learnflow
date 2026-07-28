package articlehttp

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) listAllArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	status := articledomain.ArticleStatus(r.URL.Query().Get("status"))
	if status != "" && !status.Valid() {
		h.handleErrorResponse(w, r, articledomain.ErrInvalidGetType)
		return
	}

	articleList, err := h.svc.GetAllArticles(ctx, status, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"article_list": articleList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
