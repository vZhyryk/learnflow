package noteshttp_test

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
	notesdomain "learnflow_backend/internal/notes/domain"
	noteshttp "learnflow_backend/internal/notes/transport/http"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"

	"github.com/justinas/alice"
)

// testUserID is a UUID because the request validation checks the session user id as one.
const testUserID = "11111111-1111-1111-1111-111111111111"

func withUser(r *http.Request) *http.Request {
	return r.WithContext(appcontext.WithUser(r.Context(), &authdomain.User{ID: testUserID}))
}

type mockService struct {
	createUserNotes         func(ctx context.Context, req notesdomain.CreateNotesRequest) (*notesdomain.UserNotes, error)
	getUserNotesByID        func(ctx context.Context, id, userID string) (*notesdomain.UserNotes, error)
	getUserAllNotesByUserID func(ctx context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, int, error)
	updateUserNotes         func(ctx context.Context, req notesdomain.UpdateNotesRequest) error
	deleteUserNotes         func(ctx context.Context, id, userID string) error
}

func (m *mockService) CreateUserNotes(ctx context.Context, req notesdomain.CreateNotesRequest) (*notesdomain.UserNotes, error) {
	if m.createUserNotes == nil {
		panic("mockService.CreateUserNotes not set")
	}

	return m.createUserNotes(ctx, req)
}

func (m *mockService) GetUserNotesByID(ctx context.Context, id, userID string) (*notesdomain.UserNotes, error) {
	if m.getUserNotesByID == nil {
		panic("mockService.GetUserNotesByID not set")
	}

	return m.getUserNotesByID(ctx, id, userID)
}

func (m *mockService) GetUserAllNotesByUserID(ctx context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, int, error) {
	if m.getUserAllNotesByUserID == nil {
		panic("mockService.GetUserAllNotesByUserID not set")
	}

	return m.getUserAllNotesByUserID(ctx, userID, search, params)
}

func (m *mockService) UpdateUserNotes(ctx context.Context, req notesdomain.UpdateNotesRequest) error {
	if m.updateUserNotes == nil {
		panic("mockService.UpdateUserNotes not set")
	}

	return m.updateUserNotes(ctx, req)
}

func (m *mockService) DeleteUserNotes(ctx context.Context, id, userID string) error {
	if m.deleteUserNotes == nil {
		panic("mockService.DeleteUserNotes not set")
	}

	return m.deleteUserNotes(ctx, id, userID)
}

// newHTTPFixture wires a mockService-backed mux and a request builder for a single route.
func newHTTPFixture(svc *mockService, method, path string) *testutil.HTTPFixture {
	h := noteshttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux, alice.Chain{})

	return testutil.NewHTTPFixture(mux, method, path)
}
