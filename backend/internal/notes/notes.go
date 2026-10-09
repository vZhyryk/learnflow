package notes

import (
	"learnflow_backend/internal/infrastructure/logger"
	notesdomain "learnflow_backend/internal/notes/domain"
	noteshttp "learnflow_backend/internal/notes/transport/http"
	"net/http"

	"github.com/justinas/alice"
)

// RegisterNotesRoutes registers the notes routes behind authChain.
func RegisterNotesRoutes(mux *http.ServeMux, svc notesdomain.Service, authChain alice.Chain, jsonLogger *logger.Logger) {
	notesHandler := noteshttp.NewHTTPHandler(svc, jsonLogger)
	notesHandler.RegisterRoutes(mux, authChain)
}
