package articlehttp

import (
	"net/http"
)

func (h *Handler) archiveArticle(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleAction(w, r, h.svc.ArchiveArticle, "article was successfully archived")
}
