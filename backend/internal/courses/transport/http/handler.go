package coursehttp

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/validator"
	"net/http"
)

func (h *Handler) handleSimpleAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, courseID string) error, successMsg string) {
	courseID := r.PathValue("id")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	if !validator.IsValidUUID(courseID) {
		h.handleErrorResponse(w, r, coursedomain.ErrInvalidCourseID)
		return
	}

	err := action(ctx, courseID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": successMsg}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
