package articlehttp

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) updateArticle(w http.ResponseWriter, r *http.Request) {
	var req articledomain.UpdateArticleRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	err := h.svc.UpdateArticle(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "article was successfully updated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
