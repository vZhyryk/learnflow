package notesdomain

import (
	"context"
	"learnflow_backend/internal/shared/pagination"
)

// Transactor runs fn inside a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CourseRepository checks that a course a note links to exists and is published.
type CourseRepository interface {
	CheckIfCourseExistsActiveByID(ctx context.Context, courseID string) (bool, error)
}

// ContentRepository checks that a content item a note links to exists and is published.
type ContentRepository interface {
	CheckIfContentItemExistsActiveByID(ctx context.Context, contentItemID string) (bool, error)
}

// NotesRepository persists notes; every read and write is scoped to the owning user.
type NotesRepository interface {
	CreateUserNotes(ctx context.Context, notes *UserNotes) (*UserNotes, error)
	GetUserNotesByID(ctx context.Context, id, userID string) (*UserNotes, error)
	GetUserNotesByIDForUpdate(ctx context.Context, id, userID string) (*UserNotes, error)
	GetUserAllNotesByUserID(ctx context.Context, userID, search string, params pagination.Params) ([]*UserNotes, error)
	UpdateUserNotes(ctx context.Context, notes *UserNotes) error
	DeleteUserNotes(ctx context.Context, id, userID string) error
}

// Service is the notes use-case layer.
type Service interface {
	CreateUserNotes(ctx context.Context, req CreateNotesRequest) (*UserNotes, error)
	GetUserNotesByID(ctx context.Context, id, userID string) (*UserNotes, error)
	GetUserAllNotesByUserID(ctx context.Context, userID, search string, params pagination.Params) ([]*UserNotes, error)
	UpdateUserNotes(ctx context.Context, req UpdateNotesRequest) error
	DeleteUserNotes(ctx context.Context, id, userID string) error
}
