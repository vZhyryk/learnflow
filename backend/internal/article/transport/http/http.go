package articlehttp

import (
	"net/http"

	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the article module.
type Handler struct {
	svc        articledomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the article module.
func NewHTTPHandler(svc articledomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all article HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, staticChain, adminChain alice.Chain) {
	mux.Handle("GET /api/v1/articles", staticChain.ThenFunc(h.listArticles))
	mux.Handle("GET /api/v1/articles/{slug}", staticChain.ThenFunc(h.getArticleBySlug))

	mux.Handle("GET /api/v1/admin/articles", adminChain.ThenFunc(h.listAllArticles))
	mux.Handle("POST /api/v1/admin/articles", adminChain.ThenFunc(h.createArticle))
	mux.Handle("PUT /api/v1/admin/articles", adminChain.ThenFunc(h.updateArticle))
	mux.Handle("PUT /api/v1/admin/articles/{id}/publish", adminChain.ThenFunc(h.publishArticle))
	mux.Handle("PUT /api/v1/admin/articles/{id}/archive", adminChain.ThenFunc(h.archiveArticle))
	mux.Handle("DELETE /api/v1/admin/articles/{id}", adminChain.ThenFunc(h.deleteArticle))
}
