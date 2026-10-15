package notesrepository

import (
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"strconv"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeNote(n int) *notesdomain.UserNotes {
	resourceType := notesdomain.CourseResourceType
	resourceID := "22222222-2222-2222-2222-222222222222"
	description := "description " + strconv.Itoa(n)
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	return &notesdomain.UserNotes{
		ID:           "note-" + strconv.Itoa(n),
		UserID:       "user-1",
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Title:        "title " + strconv.Itoa(n),
		Description:  &description,
		Body:         "body " + strconv.Itoa(n),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// fakeNoteScan simulates rows.Scan populating a note in the column order of notesColumns, matching scanUserNotes.
func fakeNoteScan(note *notesdomain.UserNotes) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = note.ID
		*testutil.CastStr(dest[1], 1) = note.UserID
		*testutil.CastEnum[*notesdomain.ResourceType](dest[2], 2) = note.ResourceType
		*testutil.CastPtrStr(dest[3], 3) = note.ResourceID
		*testutil.CastStr(dest[4], 4) = note.Title
		*testutil.CastPtrStr(dest[5], 5) = note.Description
		*testutil.CastStr(dest[6], 6) = note.Body
		*testutil.CastPtrTime(dest[7], 7) = note.DeletedAt
		*testutil.CastTime(dest[8], 8) = note.CreatedAt
		*testutil.CastTime(dest[9], 9) = note.UpdatedAt

		return nil
	}
}
