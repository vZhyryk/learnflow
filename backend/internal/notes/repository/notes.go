package notesrepository

import (
	"context"
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"strings"

	"github.com/jackc/pgx/v5"
)

const notesResourcePairingCheckConstraint = "user_notes_resource_check"

// likeEscaper makes %, _ and \ in a search term match literally inside ILIKE.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// CreateUserNotes implements notesdomain.NotesRepository.
func (rep *Repository) CreateUserNotes(ctx context.Context, notes *notesdomain.UserNotes) (*notesdomain.UserNotes, error) {
	result, err := scanUserNotes(rep.QueryRunner(ctx).QueryRow(ctx, createUserNotesSQL, notes.UserID, notes.ResourceType, notes.ResourceID, notes.Title, notes.Description, notes.Body))
	if db.IsCheckViolation(err, notesResourcePairingCheckConstraint) {
		return nil, notesdomain.ErrResourceDataMisMatch
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateUserNotes: %w", err)
	}

	return result, nil
}

// GetUserNotesByID implements notesdomain.NotesRepository.
func (rep *Repository) GetUserNotesByID(ctx context.Context, id, userID string) (*notesdomain.UserNotes, error) {
	result, err := scanUserNotes(rep.QueryRunner(ctx).QueryRow(ctx, getUserNotesByIDSQL, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notesdomain.ErrNoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserNotesByID: %w", err)
	}

	return result, nil
}

// GetUserAllNotesByUserID returns a page of the user's notes; a non-empty search matches title or body case-insensitively.
func (rep *Repository) GetUserAllNotesByUserID(ctx context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, error) {
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository, getUserAllNotesByUserIDSQL, "GetUserAllNotesByUserID", &params, scanUserNotes, []any{userID, likeEscaper.Replace(search)})
}

// UpdateUserNotes implements notesdomain.NotesRepository.
func (rep *Repository) UpdateUserNotes(ctx context.Context, notes *notesdomain.UserNotes) error {
	result, err := rep.QueryRunner(ctx).Exec(ctx, updateUserNotesSQL, notes.Title, notes.Description, notes.Body, notes.ID, notes.UserID)
	if err != nil {
		return fmt.Errorf("repository.UpdateUserNotes: %w", err)
	}

	if result.RowsAffected() == 0 {
		return notesdomain.ErrNoteNotFound
	}

	return nil
}

// DeleteUserNotes implements notesdomain.NotesRepository.
func (rep *Repository) DeleteUserNotes(ctx context.Context, id, userID string) error {
	result, err := rep.QueryRunner(ctx).Exec(ctx, deleteUserNotesSQL, id, userID)
	if err != nil {
		return fmt.Errorf("repository.DeleteUserNotes: %w", err)
	}

	if result.RowsAffected() == 0 {
		return notesdomain.ErrNoteNotFound
	}

	return nil
}
