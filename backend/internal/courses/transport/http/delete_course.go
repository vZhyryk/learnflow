package coursehttp

import (
	"net/http"
)

func (h *Handler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.DeleteCourse, "course was successfully deleted")
}
