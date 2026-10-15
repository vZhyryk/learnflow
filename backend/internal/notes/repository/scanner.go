package notesrepository

import (
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanUserNotes(row repository.RowScanner) (*notesdomain.UserNotes, error) {
	notes := &notesdomain.UserNotes{}
	err := row.Scan(
		&notes.ID,
		&notes.UserID,
		&notes.ResourceType,
		&notes.ResourceID,
		&notes.Title,
		&notes.Description,
		&notes.Body,
		&notes.DeletedAt,
		&notes.CreatedAt,
		&notes.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return notes, nil
}
