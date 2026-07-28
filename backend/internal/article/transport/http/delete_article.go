package articlehttp

import (
	"net/http"
)

func (h *Handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.DeleteArticle, "article was successfully deleted")
}
