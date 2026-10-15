package contenthttp

import (
	"net/http"
)

func (h *Handler) archiveContentItem(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.ArchiveContentItem, "content was successfully archived")
}
