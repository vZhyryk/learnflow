package coursehttp

import (
	"net/http"
)

func (h *Handler) publishCourse(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.PublishCourse, "course was successfully published")
}
