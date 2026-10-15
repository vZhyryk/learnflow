package article

import (
	articledomain "learnflow_backend/internal/article/domain"
	articlehttp "learnflow_backend/internal/article/transport/http"
	"learnflow_backend/internal/infrastructure/logger"
	"net/http"

	"github.com/justinas/alice"
)

// RegisterArticleRoutes wires the article module's HTTP handler onto mux.
func RegisterArticleRoutes(mux *http.ServeMux, svc articledomain.Service, staticChain, adminChain alice.Chain, jsonLogger *logger.Logger) {
	articleHandler := articlehttp.NewHTTPHandler(svc, jsonLogger)
	articleHandler.RegisterRoutes(mux, staticChain, adminChain)
}
