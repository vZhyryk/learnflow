package contenthttp

import (
	"net/http"
)

func (h *Handler) deleteContentItem(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.DeleteContentItem, "content was successfully deleted")
}
