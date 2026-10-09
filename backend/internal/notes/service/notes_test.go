package notesservice

import (
	"context"
	"errors"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testUserID     = "11111111-1111-1111-1111-111111111111"
	testNoteID     = "22222222-2222-2222-2222-222222222222"
	testResourceID = "33333333-3333-3333-3333-333333333333"
)

func linkedRequest(resourceType notesdomain.ResourceType) notesdomain.CreateNotesRequest {
	resourceID := testResourceID

	return notesdomain.CreateNotesRequest{
		UserID: testUserID, ResourceType: &resourceType, ResourceID: &resourceID, Title: "title", Body: "body",
	}
}

func TestCreateUserNotesPersonal(t *testing.T) {
	Convey("Given a note without a linked resource", t, func() {
		var created *notesdomain.UserNotes
		notesRepo := &mockNotesRepo{createUserNotes: func(_ context.Context, notes *notesdomain.UserNotes) (*notesdomain.UserNotes, error) {
			created = notes
			return notes, nil
		}}
		// The course and content mocks have no functions set: calling either would panic.
		srv := newTestService(notesRepo, &mockCourseRepo{}, &mockContentRepo{})
		description := "desc"

		Convey("When it is created, no resource is checked and every field, including the owner, is persisted", func() {
			req := notesdomain.CreateNotesRequest{UserID: testUserID, Title: "title", Description: &description, Body: "body"}

			got, err := srv.CreateUserNotes(context.Background(), req)

			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
			So(created, ShouldResemble, &notesdomain.UserNotes{UserID: testUserID, Title: "title", Description: &description, Body: "body"})
		})

		Convey("When the repository fails, the error is wrapped with the service method name", func() {
			notesRepo.createUserNotes = func(_ context.Context, _ *notesdomain.UserNotes) (*notesdomain.UserNotes, error) {
				return nil, testutil.ErrDBUnexpected
			}

			_, err := srv.CreateUserNotes(context.Background(), notesdomain.CreateNotesRequest{UserID: testUserID, Title: "t", Body: "b"})

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.CreateUserNotes")
		})
	})
}

func TestCreateUserNotesLinked(t *testing.T) {
	Convey("Given a note linked to a course or content item", t, func() {
		var created *notesdomain.UserNotes
		var checkedCourse, checkedContent string
		courseExists, contentExists := true, true
		var checkErr error
		notesRepo := &mockNotesRepo{createUserNotes: func(_ context.Context, notes *notesdomain.UserNotes) (*notesdomain.UserNotes, error) {
			created = notes
			return notes, nil
		}}
		courseRepo := &mockCourseRepo{checkIfCourseExistsActiveByID: func(_ context.Context, id string) (bool, error) {
			checkedCourse = id
			return courseExists, checkErr
		}}
		contentRepo := &mockContentRepo{checkIfContentItemExistsActiveByID: func(_ context.Context, id string) (bool, error) {
			checkedContent = id
			return contentExists, checkErr
		}}
		srv := newTestService(notesRepo, courseRepo, contentRepo)

		Convey("When the course is published, only the course repository is asked and the note is created", func() {
			_, err := srv.CreateUserNotes(context.Background(), linkedRequest(notesdomain.CourseResourceType))

			So(err, ShouldBeNil)
			So(checkedCourse, ShouldEqual, testResourceID)
			So(checkedContent, ShouldBeEmpty)
			So(created, ShouldNotBeNil)
		})

		Convey("When the content item is published, only the content repository is asked and the note is created", func() {
			_, err := srv.CreateUserNotes(context.Background(), linkedRequest(notesdomain.ContentResourceType))

			So(err, ShouldBeNil)
			So(checkedContent, ShouldEqual, testResourceID)
			So(checkedCourse, ShouldBeEmpty)
			So(created, ShouldNotBeNil)
		})

		Convey("When the course does not exist or is unpublished, it is ErrInvalidResourceID and nothing is stored", func() {
			courseExists = false

			_, err := srv.CreateUserNotes(context.Background(), linkedRequest(notesdomain.CourseResourceType))

			So(errors.Is(err, notesdomain.ErrInvalidResourceID), ShouldBeTrue)
			So(created, ShouldBeNil)
		})

		Convey("When the content item does not exist or is unpublished, it is ErrInvalidResourceID and nothing is stored", func() {
			contentExists = false

			_, err := srv.CreateUserNotes(context.Background(), linkedRequest(notesdomain.ContentResourceType))

			So(errors.Is(err, notesdomain.ErrInvalidResourceID), ShouldBeTrue)
			So(created, ShouldBeNil)
		})

		Convey("When the existence check fails, the error is wrapped and nothing is stored", func() {
			checkErr = testutil.ErrDBUnexpected

			_, err := srv.CreateUserNotes(context.Background(), linkedRequest(notesdomain.CourseResourceType))

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.CreateUserNotes")
			So(created, ShouldBeNil)
		})
	})
}

func TestGetUserNotes(t *testing.T) {
	Convey("Given a notes service", t, func() {
		var gotID, gotUserID, gotSearch string
		var gotParams pagination.Params
		var repoErr error
		notesRepo := &mockNotesRepo{
			getUserNotesByID: func(_ context.Context, id, userID string) (*notesdomain.UserNotes, error) {
				gotID, gotUserID = id, userID
				return &notesdomain.UserNotes{ID: id}, repoErr
			},
			getUserAllNotesByUserID: func(_ context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, error) {
				gotUserID, gotSearch, gotParams = userID, search, params
				return []*notesdomain.UserNotes{{ID: "a"}, {ID: "b"}}, repoErr
			},
		}
		srv := newTestService(notesRepo, &mockCourseRepo{}, &mockContentRepo{})

		Convey("GetUserNotesByID passes the id and the owner to the repository", func() {
			got, err := srv.GetUserNotesByID(context.Background(), testNoteID, testUserID)

			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, testNoteID)
			So(gotID, ShouldEqual, testNoteID)
			So(gotUserID, ShouldEqual, testUserID)
		})

		Convey("GetUserNotesByID keeps ErrNoteNotFound recognisable", func() {
			repoErr = notesdomain.ErrNoteNotFound

			_, err := srv.GetUserNotesByID(context.Background(), testNoteID, testUserID)

			So(errors.Is(err, notesdomain.ErrNoteNotFound), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetUserNotesByID")
		})

		Convey("GetUserAllNotesByUserID passes the owner, the search term and the page to the repository", func() {
			got, err := srv.GetUserAllNotesByUserID(context.Background(), testUserID, "alpha", pagination.NewParams(2, 5))

			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(gotUserID, ShouldEqual, testUserID)
			So(gotSearch, ShouldEqual, "alpha")
			So(gotParams, ShouldResemble, pagination.NewParams(2, 5))
		})

		Convey("GetUserAllNotesByUserID wraps a repository error", func() {
			repoErr = testutil.ErrDBUnexpected

			_, err := srv.GetUserAllNotesByUserID(context.Background(), testUserID, "", pagination.NewParams(1, 10))

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetUserAllNotesByUserID")
		})
	})
}

func TestUpdateUserNotes(t *testing.T) {
	Convey("Given a notes service", t, func() {
		stored := &notesdomain.UserNotes{ID: testNoteID, UserID: testUserID, Title: "old", Body: "old body"}
		var gotLookupUserID string
		var updated *notesdomain.UserNotes
		var getErr, updateErr error
		notesRepo := &mockNotesRepo{
			getUserNotesByID: func(_ context.Context, _, userID string) (*notesdomain.UserNotes, error) {
				gotLookupUserID = userID
				return stored, getErr
			},
			updateUserNotes: func(_ context.Context, notes *notesdomain.UserNotes) error {
				updated = notes
				return updateErr
			},
		}
		srv := newTestService(notesRepo, &mockCourseRepo{}, &mockContentRepo{})
		newTitle := "new"

		Convey("When the owner updates a field, only that field changes and the lookup is scoped to the owner", func() {
			err := srv.UpdateUserNotes(context.Background(), notesdomain.UpdateNotesRequest{ID: testNoteID, UserID: testUserID, Title: &newTitle})

			So(err, ShouldBeNil)
			So(gotLookupUserID, ShouldEqual, testUserID)
			So(updated.Title, ShouldEqual, "new")
			So(updated.Body, ShouldEqual, "old body")
		})

		Convey("When the note is not found for this owner, nothing is updated", func() {
			getErr = notesdomain.ErrNoteNotFound

			err := srv.UpdateUserNotes(context.Background(), notesdomain.UpdateNotesRequest{ID: testNoteID, UserID: testUserID, Title: &newTitle})

			So(errors.Is(err, notesdomain.ErrNoteNotFound), ShouldBeTrue)
			So(updated, ShouldBeNil)
		})

		Convey("When the repository update fails, the error is wrapped", func() {
			updateErr = testutil.ErrDBUnexpected

			err := srv.UpdateUserNotes(context.Background(), notesdomain.UpdateNotesRequest{ID: testNoteID, UserID: testUserID, Title: &newTitle})

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.UpdateUserNotes")
		})
	})
}

func TestDeleteUserNotes(t *testing.T) {
	Convey("Given a notes service", t, func() {
		var gotID, gotUserID string
		var repoErr error
		srv := newTestService(&mockNotesRepo{deleteUserNotes: func(_ context.Context, id, userID string) error {
			gotID, gotUserID = id, userID
			return repoErr
		}}, &mockCourseRepo{}, &mockContentRepo{})

		Convey("When the owner deletes a note, the id and the owner reach the repository", func() {
			So(srv.DeleteUserNotes(context.Background(), testNoteID, testUserID), ShouldBeNil)
			So(gotID, ShouldEqual, testNoteID)
			So(gotUserID, ShouldEqual, testUserID)
		})

		Convey("When the repository reports not found, it stays recognisable and is wrapped", func() {
			repoErr = notesdomain.ErrNoteNotFound

			err := srv.DeleteUserNotes(context.Background(), testNoteID, testUserID)

			So(errors.Is(err, notesdomain.ErrNoteNotFound), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.DeleteUserNotes")
		})
	})
}
