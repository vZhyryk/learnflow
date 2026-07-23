package coursehttp

import (
	"net/http"
)

func (h *Handler) archiveCourse(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.ArchiveCourse, "course was successfully archived")
}
