package noteshttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	notesdomain "learnflow_backend/internal/notes/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/validator"
	"net/http"
)

// notePathID returns the {id} path value; a value that is not a UUID cannot match any note, so it is answered with 404.
func (h *Handler) notePathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	noteID := r.PathValue("id")
	if validator.IsValidUUID(noteID) {
		return noteID, true
	}

	h.handleErrorRespond(r, "note_id_not_uuid", func() error {
		return helpers.NotFoundResponse(w)
	})

	return "", false
}

func (h *Handler) createUserNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	var req notesdomain.CreateNotesRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
	}) {
		return
	}

	note, err := h.svc.CreateUserNotes(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"note": note}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) getUserNotesByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	noteID, ok := h.notePathID(w, r)
	if !ok {
		return
	}

	notes, err := h.svc.GetUserNotesByID(ctx, noteID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"notes": notes}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) getUserAllNotesByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	filter := r.URL.Query().Get("filter")

	notesList, err := h.svc.GetUserAllNotesByUserID(ctx, user.ID, filter, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"note_list": notesList}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) updateUserNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	noteID, ok := h.notePathID(w, r)
	if !ok {
		return
	}

	var req notesdomain.UpdateNotesRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() {
		req.UserID = user.ID
		req.ID = noteID
	}) {
		return
	}

	err := h.svc.UpdateUserNotes(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "note was successfully updated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) deleteUserNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	noteID, ok := h.notePathID(w, r)
	if !ok {
		return
	}

	err := h.svc.DeleteUserNotes(ctx, noteID, user.ID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "note was successfully deleted"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
