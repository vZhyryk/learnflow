package articlehttp

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) createArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req articledomain.CreateArticleRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.CreatedByUserID = user.ID
	}) {
		return
	}

	articleID, err := h.svc.CreateArticle(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"article_id": articleID}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
