package articlehttp

import (
	"context"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/validator"
	"net/http"
)

func (h *Handler) handleSimpleAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, articleID, userID string) error, successMsg string) {
	articleID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(articleID) {
		h.handleErrorResponse(w, r, articledomain.ErrInvalidArticleID)
		return
	}

	err := action(ctx, articleID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": successMsg}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
