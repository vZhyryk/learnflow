package noteshttp_test

import (
	"context"
	"fmt"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	noteID     = "22222222-2222-2222-2222-222222222222"
	resourceID = "33333333-3333-3333-3333-333333333333"
	notesPath  = "/api/v1/notes"
)

func TestCreateUserNotesRoute(t *testing.T) {
	Convey("POST "+notesPath, t, func() {
		var gotReq notesdomain.CreateNotesRequest
		var svcErr error
		svc := &mockService{createUserNotes: func(_ context.Context, req notesdomain.CreateNotesRequest) (*notesdomain.UserNotes, error) {
			gotReq = req
			return &notesdomain.UserNotes{ID: noteID, UserID: req.UserID, Title: req.Title, Body: req.Body}, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodPost, notesPath)
		serve := func(body string) int { return testutil.ServeHTTP(f.Mux, withUser(f.NewReq(body, nil))).Code }

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() { testutil.ServeHTTP(f.Mux, f.NewReq(`{"title":"t","body":"b"}`, nil)) }, ShouldPanic)
		})

		Convey("A valid note → 201 with the created note, and the owner comes from the session", func() {
			w := testutil.ServeHTTP(f.Mux, withUser(f.NewReq(`{"title":"t","body":"b"}`, nil)))

			So(w.Code, ShouldEqual, http.StatusCreated)
			note, ok := testutil.DecodeBody(t, w.Body.Bytes())["note"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(note["id"], ShouldEqual, noteID)
			So(gotReq.UserID, ShouldEqual, testUserID)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(serve(`{"title":"t","body":"b"}`), ShouldEqual, http.StatusInternalServerError)
		})

		Convey("The success response write fails → does not panic", func() {
			So(func() { f.Mux.ServeHTTP(&testutil.ErrWriter{}, withUser(f.NewReq(`{"title":"t","body":"b"}`, nil))) }, ShouldNotPanic)
		})
	})
}

func TestCreateUserNotesRouteRejections(t *testing.T) {
	Convey("POST "+notesPath+" — rejected requests", t, func() {
		var gotReq notesdomain.CreateNotesRequest
		var svcErr error
		svc := &mockService{createUserNotes: func(_ context.Context, req notesdomain.CreateNotesRequest) (*notesdomain.UserNotes, error) {
			gotReq = req
			return nil, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodPost, notesPath)
		serve := func(body string) int { return testutil.ServeHTTP(f.Mux, withUser(f.NewReq(body, nil))).Code }

		Convey("A user_id sent by the client is rejected, so it cannot pick the owner", func() {
			So(serve(`{"title":"t","body":"b","user_id":"someone-else"}`), ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid input → 400 and never reaches the service", func() {
			for _, body := range []string{
				`{`,
				`{"body":"b"}`,
				`{"title":"t"}`,
				`{"title":"t","body":"b","resource_type":"course"}`,
				`{"title":"t","body":"b","resource_type":"video","resource_id":"` + resourceID + `"}`,
				`{"title":"t","body":"b","resource_type":"course","resource_id":"nope"}`,
			} {
				So(fmt.Sprintf("%s → %d", body, serve(body)), ShouldEqual, fmt.Sprintf("%s → %d", body, http.StatusBadRequest))
			}
			So(gotReq.Title, ShouldBeEmpty)
		})

		Convey("A link to a missing or unpublished resource → 422 with a fixed message", func() {
			svcErr = fmt.Errorf("service.CreateUserNotes: %w", notesdomain.ErrInvalidResourceID)

			w := testutil.ServeHTTP(f.Mux, withUser(f.NewReq(`{"title":"t","body":"b"}`, nil)))

			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
			So(w.Body.String(), ShouldContainSubstring, notesdomain.ErrInvalidResourceID.Error())
			So(w.Body.String(), ShouldNotContainSubstring, "service.")
		})
	})
}

func TestGetUserNotesByIDRoute(t *testing.T) {
	Convey("GET "+notesPath+"/{id}", t, func() {
		var called bool
		var gotID, gotUserID string
		var svcErr error
		svc := &mockService{getUserNotesByID: func(_ context.Context, id, userID string) (*notesdomain.UserNotes, error) {
			called, gotID, gotUserID = true, id, userID
			return &notesdomain.UserNotes{ID: id, UserID: userID, Title: "t"}, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodGet, notesPath+"/"+noteID)

		Convey("A note of the caller → 200, looked up by id and owner", func() {
			w := testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil)))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotID, ShouldEqual, noteID)
			So(gotUserID, ShouldEqual, testUserID)
			So(testutil.DecodeBody(t, w.Body.Bytes()), ShouldContainKey, "note")
		})

		Convey("A note of another user or a missing note → 404", func() {
			svcErr = fmt.Errorf("service.GetUserNotesByID: %w", notesdomain.ErrNoteNotFound)

			So(testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil))).Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("An id that is not a UUID → 404 and the service is never called", func() {
			bad := newHTTPFixture(svc, http.MethodGet, notesPath+"/not-a-uuid")

			So(testutil.ServeHTTP(bad.Mux, withUser(bad.NewReq("", nil))).Code, ShouldEqual, http.StatusNotFound)
			So(called, ShouldBeFalse)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil))).Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestGetUserAllNotesRoute(t *testing.T) {
	Convey("GET "+notesPath, t, func() {
		var gotUserID, gotSearch string
		var gotParams pagination.Params
		var svcErr error
		svc := &mockService{getUserAllNotesByUserID: func(_ context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, int, error) {
			gotUserID, gotSearch, gotParams = userID, search, params
			return []*notesdomain.UserNotes{{ID: "a"}, {ID: "b"}}, 7, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodGet, notesPath)

		Convey("The list is scoped to the caller, with the search term and the page from the query", func() {
			w := testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", map[string]string{"filter": "alpha", "page": "2", "page_size": "5"})))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotUserID, ShouldEqual, testUserID)
			So(gotSearch, ShouldEqual, "alpha")
			So(gotParams, ShouldResemble, pagination.NewParams(2, 5))
			body := testutil.DecodeBody(t, w.Body.Bytes())
			list, ok := body["notes"].([]any)
			So(ok, ShouldBeTrue)
			So(list, ShouldHaveLength, 2)
			So(body["total"], ShouldEqual, 7)
		})

		Convey("A filter of exactly 100 characters is accepted", func() {
			w := testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", map[string]string{"filter": strings.Repeat("я", 100)})))

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("A filter longer than 100 characters → 422 and the service is not called", func() {
			w := testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", map[string]string{"filter": strings.Repeat("я", 101)})))

			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
			So(gotUserID, ShouldBeEmpty)
		})

		Convey("Without a query the search is empty", func() {
			testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil)))

			So(gotSearch, ShouldBeEmpty)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil))).Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestUpdateUserNotesRoute(t *testing.T) {
	Convey("PUT "+notesPath+"/{id}", t, func() {
		var called bool
		var gotReq notesdomain.UpdateNotesRequest
		var svcErr error
		svc := &mockService{updateUserNotes: func(_ context.Context, req notesdomain.UpdateNotesRequest) error {
			called, gotReq = true, req
			return svcErr
		}}
		f := newHTTPFixture(svc, http.MethodPut, notesPath+"/"+noteID)
		serve := func(body string) int { return testutil.ServeHTTP(f.Mux, withUser(f.NewReq(body, nil))).Code }

		Convey("A valid update → 200, with the id from the path and the owner from the session", func() {
			So(serve(`{"title":"new"}`), ShouldEqual, http.StatusOK)
			So(gotReq.ID, ShouldEqual, noteID)
			So(gotReq.UserID, ShouldEqual, testUserID)
			So(*gotReq.Title, ShouldEqual, "new")
		})

		Convey("Invalid input → 400 and never reaches the service", func() {
			for _, body := range []string{`{`, `{"title":" "}`, `{"body":""}`, `{"resource_type":"course"}`, `{"id":"x"}`, `{"user_id":"x"}`} {
				So(fmt.Sprintf("%s → %d", body, serve(body)), ShouldEqual, fmt.Sprintf("%s → %d", body, http.StatusBadRequest))
			}
			So(called, ShouldBeFalse)
		})

		Convey("An id that is not a UUID → 404 and the service is never called", func() {
			bad := newHTTPFixture(svc, http.MethodPut, notesPath+"/xyz")

			So(testutil.ServeHTTP(bad.Mux, withUser(bad.NewReq(`{"title":"new"}`, nil))).Code, ShouldEqual, http.StatusNotFound)
			So(called, ShouldBeFalse)
		})

		Convey("A note of another user or a missing note → 404", func() {
			svcErr = fmt.Errorf("service.UpdateUserNotes: %w", notesdomain.ErrNoteNotFound)
			So(serve(`{"title":"new"}`), ShouldEqual, http.StatusNotFound)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(serve(`{"title":"new"}`), ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestDeleteUserNotesRoute(t *testing.T) {
	Convey("DELETE "+notesPath+"/{id}", t, func() {
		var called bool
		var gotID, gotUserID string
		var svcErr error
		svc := &mockService{deleteUserNotes: func(_ context.Context, id, userID string) error {
			called, gotID, gotUserID = true, id, userID
			return svcErr
		}}
		f := newHTTPFixture(svc, http.MethodDelete, notesPath+"/"+noteID)

		Convey("The caller's note → 200, deleted by id and owner", func() {
			So(testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil))).Code, ShouldEqual, http.StatusOK)
			So(gotID, ShouldEqual, noteID)
			So(gotUserID, ShouldEqual, testUserID)
		})

		Convey("An id that is not a UUID → 404 and the service is never called", func() {
			bad := newHTTPFixture(svc, http.MethodDelete, notesPath+"/xyz")

			So(testutil.ServeHTTP(bad.Mux, withUser(bad.NewReq("", nil))).Code, ShouldEqual, http.StatusNotFound)
			So(called, ShouldBeFalse)
		})

		Convey("A note of another user, a missing or an already deleted note → 404", func() {
			svcErr = fmt.Errorf("service.DeleteUserNotes: %w", notesdomain.ErrNoteNotFound)
			So(testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil))).Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(testutil.ServeHTTP(f.Mux, withUser(f.NewReq("", nil))).Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
