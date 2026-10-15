package contenthttp

import (
	"context"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/validator"
	"net/http"
)

func (h *Handler) handleSimpleAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, contentItemID, userID string) error, successMsg string) {
	contentItemID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(contentItemID) {
		h.handleErrorResponse(w, r, contentdomain.ErrInvalidContentItemID)
		return
	}

	err := action(ctx, contentItemID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": successMsg}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
