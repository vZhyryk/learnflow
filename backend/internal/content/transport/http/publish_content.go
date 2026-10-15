package contenthttp

import (
	"net/http"
)

func (h *Handler) publishContentItem(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.PublishContentItem, "content was successfully published")
}
