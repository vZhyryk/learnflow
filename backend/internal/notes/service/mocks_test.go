package notesservice

import (
	"context"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

type mockNotesRepo struct {
	createUserNotes         func(ctx context.Context, notes *notesdomain.UserNotes) (*notesdomain.UserNotes, error)
	getUserNotesByID        func(ctx context.Context, id, userID string) (*notesdomain.UserNotes, error)
	getUserAllNotesByUserID func(ctx context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, error)
	updateUserNotes         func(ctx context.Context, notes *notesdomain.UserNotes) error
	deleteUserNotes         func(ctx context.Context, id, userID string) error
}

func (m *mockNotesRepo) CreateUserNotes(ctx context.Context, notes *notesdomain.UserNotes) (*notesdomain.UserNotes, error) {
	if m.createUserNotes == nil {
		panic("mockNotesRepo.CreateUserNotes not set")
	}

	return m.createUserNotes(ctx, notes)
}

func (m *mockNotesRepo) GetUserNotesByID(ctx context.Context, id, userID string) (*notesdomain.UserNotes, error) {
	if m.getUserNotesByID == nil {
		panic("mockNotesRepo.GetUserNotesByID not set")
	}

	return m.getUserNotesByID(ctx, id, userID)
}

func (m *mockNotesRepo) GetUserAllNotesByUserID(ctx context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, error) {
	if m.getUserAllNotesByUserID == nil {
		panic("mockNotesRepo.GetUserAllNotesByUserID not set")
	}

	return m.getUserAllNotesByUserID(ctx, userID, search, params)
}

func (m *mockNotesRepo) UpdateUserNotes(ctx context.Context, notes *notesdomain.UserNotes) error {
	if m.updateUserNotes == nil {
		panic("mockNotesRepo.UpdateUserNotes not set")
	}

	return m.updateUserNotes(ctx, notes)
}

func (m *mockNotesRepo) DeleteUserNotes(ctx context.Context, id, userID string) error {
	if m.deleteUserNotes == nil {
		panic("mockNotesRepo.DeleteUserNotes not set")
	}

	return m.deleteUserNotes(ctx, id, userID)
}

type mockCourseRepo struct {
	checkIfCourseExistsActiveByID func(ctx context.Context, courseID string) (bool, error)
}

func (m *mockCourseRepo) CheckIfCourseExistsActiveByID(ctx context.Context, courseID string) (bool, error) {
	if m.checkIfCourseExistsActiveByID == nil {
		panic("mockCourseRepo.CheckIfCourseExistsActiveByID not set")
	}

	return m.checkIfCourseExistsActiveByID(ctx, courseID)
}

type mockContentRepo struct {
	checkIfContentItemExistsActiveByID func(ctx context.Context, contentItemID string) (bool, error)
}

func (m *mockContentRepo) CheckIfContentItemExistsActiveByID(ctx context.Context, contentItemID string) (bool, error) {
	if m.checkIfContentItemExistsActiveByID == nil {
		panic("mockContentRepo.CheckIfContentItemExistsActiveByID not set")
	}

	return m.checkIfContentItemExistsActiveByID(ctx, contentItemID)
}

func newTestService(notesRepo *mockNotesRepo, courseRepo *mockCourseRepo, contentRepo *mockContentRepo) *Service {
	return New(notesRepo, courseRepo, contentRepo, testutil.NoopTransactor{})
}
