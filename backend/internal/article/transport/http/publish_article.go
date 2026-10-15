package articlehttp

import (
	"net/http"
)

func (h *Handler) publishArticle(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.PublishArticle, "article was successfully published")
}
