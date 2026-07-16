package coursehttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) getCourseBySlug(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("slug")
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	course, err := h.svc.GetCourseBySlug(ctx, courseID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"course": course}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
