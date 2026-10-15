package noteshttp

import (
	notesdomain "learnflow_backend/internal/notes/domain"
	"net/http"

	"learnflow_backend/internal/infrastructure/logger"

	"github.com/justinas/alice"
)

// Handler wires HTTP routes for the notes module.
type Handler struct {
	svc        notesdomain.Service
	jsonLogger *logger.Logger
}

// NewHTTPHandler returns a new Handler for the notes module.
func NewHTTPHandler(svc notesdomain.Service, jsonLogger *logger.Logger) *Handler {
	return &Handler{
		svc:        svc,
		jsonLogger: jsonLogger,
	}
}

// RegisterRoutes registers all notes HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authChain alice.Chain) {
	mux.Handle("POST /api/v1/notes", authChain.ThenFunc(h.createUserNotes))
	mux.Handle("GET /api/v1/notes/{id}", authChain.ThenFunc(h.getUserNotesByID))
	mux.Handle("GET /api/v1/notes", authChain.ThenFunc(h.getUserAllNotesByUserID))
	mux.Handle("PUT /api/v1/notes/{id}", authChain.ThenFunc(h.updateUserNotes))
	mux.Handle("DELETE /api/v1/notes/{id}", authChain.ThenFunc(h.deleteUserNotes))
}
