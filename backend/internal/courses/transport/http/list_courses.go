package coursehttp

import (
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) listCourses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	courseList, err := h.svc.GetAllCourses(ctx, coursedomain.PublishedStatus, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"course_list": courseList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
