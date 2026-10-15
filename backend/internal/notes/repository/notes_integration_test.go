//go:build integration

package notesrepository

import (
	"context"
	"errors"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func newIntegrationRepo(tx pgx.Tx) *Repository {
	return &Repository{repository.BaseRepository{DB: tx}}
}

func createNote(ctx context.Context, repo *Repository, userID, title, body string) *notesdomain.UserNotes {
	note, err := repo.CreateUserNotes(ctx, &notesdomain.UserNotes{UserID: userID, Title: title, Body: body})
	So(err, ShouldBeNil)

	return note
}

func TestCreateUserNotes_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a notes repository backed by real Postgres", t, func() {
		Convey("A personal note is stored and returned with its generated id and timestamps", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				description := "desc"

				got, err := repo.CreateUserNotes(ctx, &notesdomain.UserNotes{UserID: userID, Title: "title", Description: &description, Body: "body"})

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.UserID, ShouldEqual, userID)
				So(got.ResourceType, ShouldBeNil)
				So(got.ResourceID, ShouldBeNil)
				So(*got.Description, ShouldEqual, "desc")
				So(got.CreatedAt.IsZero(), ShouldBeFalse)
				So(got.DeletedAt, ShouldBeNil)
			})
		})

		Convey("A note linked to a resource keeps the link", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				resourceType, resourceID := notesdomain.CourseResourceType, "33333333-3333-3333-3333-333333333333"

				got, err := repo.CreateUserNotes(ctx, &notesdomain.UserNotes{
					UserID: userID, ResourceType: &resourceType, ResourceID: &resourceID, Title: "title", Body: "body",
				})

				So(err, ShouldBeNil)
				So(*got.ResourceType, ShouldEqual, notesdomain.CourseResourceType)
				So(*got.ResourceID, ShouldEqual, resourceID)
			})
		})

		Convey("A resource type without an id violates the pairing CHECK → ErrResourceDataMisMatch", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				resourceType := notesdomain.CourseResourceType

				_, err := repo.CreateUserNotes(ctx, &notesdomain.UserNotes{UserID: userID, ResourceType: &resourceType, Title: "title", Body: "body"})

				So(errors.Is(err, notesdomain.ErrResourceDataMisMatch), ShouldBeTrue)
			})
		})
	})
}

func TestNotesOwnership_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a note owned by one user and a second user", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := newIntegrationRepo(tx)
			owner := testutil.InsertRandomTestUser(t, tx)
			stranger := testutil.InsertRandomTestUser(t, tx)
			note := createNote(ctx, repo, owner, "private", "secret body")

			Convey("The owner reads the note", func() {
				got, err := repo.GetUserNotesByID(ctx, note.ID, owner)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, note.ID)
			})

			Convey("Another user cannot read, update or delete it — every call looks like a missing note", func() {
				_, getErr := repo.GetUserNotesByID(ctx, note.ID, stranger)
				updateErr := repo.UpdateUserNotes(ctx, &notesdomain.UserNotes{ID: note.ID, UserID: stranger, Title: "hijacked", Body: "x"})
				deleteErr := repo.DeleteUserNotes(ctx, note.ID, stranger)

				So(errors.Is(getErr, notesdomain.ErrNoteNotFound), ShouldBeTrue)
				So(errors.Is(updateErr, notesdomain.ErrNoteNotFound), ShouldBeTrue)
				So(errors.Is(deleteErr, notesdomain.ErrNoteNotFound), ShouldBeTrue)

				still, err := repo.GetUserNotesByID(ctx, note.ID, owner)
				So(err, ShouldBeNil)
				So(still.Title, ShouldEqual, "private")
			})

			Convey("An unknown id is a missing note", func() {
				_, err := repo.GetUserNotesByID(ctx, "00000000-0000-0000-0000-000000000000", owner)

				So(errors.Is(err, notesdomain.ErrNoteNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateAndDeleteUserNotes_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a note owned by a user", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := newIntegrationRepo(tx)
			owner := testutil.InsertRandomTestUser(t, tx)
			note := createNote(ctx, repo, owner, "old", "old body")

			Convey("An update changes title, description and body and moves updated_at", func() {
				description := "new desc"
				So(repo.UpdateUserNotes(ctx, &notesdomain.UserNotes{ID: note.ID, UserID: owner, Title: "new", Description: &description, Body: "new body"}), ShouldBeNil)

				got, err := repo.GetUserNotesByID(ctx, note.ID, owner)
				So(err, ShouldBeNil)
				So(got.Title, ShouldEqual, "new")
				So(*got.Description, ShouldEqual, "new desc")
				So(got.Body, ShouldEqual, "new body")
				So(got.UpdatedAt.Before(got.CreatedAt), ShouldBeFalse)
			})

			Convey("A delete is soft: the row stays but the note is gone for every read and write", func() {
				So(repo.DeleteUserNotes(ctx, note.ID, owner), ShouldBeNil)

				var deletedAtSet bool
				So(tx.QueryRow(ctx, `SELECT deleted_at IS NOT NULL FROM user_notes WHERE id = $1`, note.ID).Scan(&deletedAtSet), ShouldBeNil)
				So(deletedAtSet, ShouldBeTrue)

				_, getErr := repo.GetUserNotesByID(ctx, note.ID, owner)
				So(errors.Is(getErr, notesdomain.ErrNoteNotFound), ShouldBeTrue)
				So(errors.Is(repo.DeleteUserNotes(ctx, note.ID, owner), notesdomain.ErrNoteNotFound), ShouldBeTrue)
				So(errors.Is(repo.UpdateUserNotes(ctx, &notesdomain.UserNotes{ID: note.ID, UserID: owner, Title: "t", Body: "b"}), notesdomain.ErrNoteNotFound), ShouldBeTrue)

				list, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "", pagination.NewParams(1, 10))
				So(err, ShouldBeNil)
				So(list, ShouldBeEmpty)
			})
		})
	})
}

func titles(notes []*notesdomain.UserNotes) []string {
	out := make([]string, 0, len(notes))
	for _, note := range notes {
		out = append(out, note.Title)
	}

	return out
}

func TestGetUserAllNotesByUserID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given three notes of one user and one note of another", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := newIntegrationRepo(tx)
			owner := testutil.InsertRandomTestUser(t, tx)
			stranger := testutil.InsertRandomTestUser(t, tx)
			createNote(ctx, repo, owner, "Alpha", "first")
			createNote(ctx, repo, owner, "Beta", "mentions ALPHA in the body")
			createNote(ctx, repo, owner, "100% done_ok", "plain")
			createNote(ctx, repo, stranger, "Alpha of a stranger", "x")
			// Rows of one transaction share now(), so give them a stable order for the newest-first check.
			_, err := tx.Exec(ctx, `UPDATE user_notes SET created_at = now() + (CASE title WHEN 'Alpha' THEN 1 WHEN 'Beta' THEN 2 ELSE 3 END) * interval '1 second' WHERE user_id = $1`, owner)
			So(err, ShouldBeNil)
			params := pagination.NewParams(1, 10)

			Convey("Without a search it lists only the caller's notes, newest first", func() {
				list, total, err := repo.GetUserAllNotesByUserID(ctx, owner, "", params)

				So(err, ShouldBeNil)
				So(titles(list), ShouldResemble, []string{"100% done_ok", "Beta", "Alpha"})
				So(total, ShouldEqual, 3)
			})

			Convey("A search matches title or body case-insensitively and never other users' notes", func() {
				list, total, err := repo.GetUserAllNotesByUserID(ctx, owner, "alpha", params)

				So(err, ShouldBeNil)
				So(titles(list), ShouldResemble, []string{"Beta", "Alpha"})
				So(total, ShouldEqual, 2)
			})

			Convey("A search with no match is an empty page", func() {
				list, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "zzz", params)

				So(err, ShouldBeNil)
				So(list, ShouldBeEmpty)
			})

			Convey("A % or _ in the search matches literally, not as a wildcard", func() {
				percent, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "%", params)
				So(err, ShouldBeNil)
				So(titles(percent), ShouldResemble, []string{"100% done_ok"})

				underscore, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "done_ok", params)
				So(err, ShouldBeNil)
				So(titles(underscore), ShouldResemble, []string{"100% done_ok"})

				wildcard, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "done_", params)
				So(err, ShouldBeNil)
				So(titles(wildcard), ShouldResemble, []string{"100% done_ok"})

				noWildcard, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "d_ne", params)
				So(err, ShouldBeNil)
				So(noWildcard, ShouldBeEmpty)
			})

			Convey("Pagination slices the newest-first list", func() {
				first, total, err := repo.GetUserAllNotesByUserID(ctx, owner, "", pagination.NewParams(1, 2))
				So(err, ShouldBeNil)
				second, _, err := repo.GetUserAllNotesByUserID(ctx, owner, "", pagination.NewParams(2, 2))
				So(err, ShouldBeNil)

				So(total, ShouldEqual, 3)
				So(titles(first), ShouldResemble, []string{"100% done_ok", "Beta"})
				So(titles(second), ShouldResemble, []string{"Alpha"})
			})
		})
	})
}
