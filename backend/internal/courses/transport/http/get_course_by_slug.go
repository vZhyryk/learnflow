package coursehttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

func (h *Handler) getCourseBySlug(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("slug")
	ctx := r.Context()
	course, err := h.svc.GetCourseBySlug(ctx, courseID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"course": course}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
