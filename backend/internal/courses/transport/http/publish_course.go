package coursehttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) publishCourse(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")

	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	err := h.svc.PublishCourse(ctx, courseID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"status": "success"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
