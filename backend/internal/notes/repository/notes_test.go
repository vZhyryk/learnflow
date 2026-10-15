package notesrepository

import (
	"context"
	"errors"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateUserNotes(t *testing.T) {
	Convey("Given a notes repository", t, func() {
		var row *testutil.MockRow
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, args ...any) pgx.Row {
				gotArgs = args
				return row
			},
		})
		note := fakeNote(1)

		Convey("When the insert succeeds, the returned row is scanned and every column is passed in order", func() {
			row = &testutil.MockRow{ScanFn: fakeNoteScan(note)}

			got, err := repo.CreateUserNotes(context.Background(), note)

			So(err, ShouldBeNil)
			So(got, ShouldResemble, note)
			So(gotArgs, ShouldResemble, []any{note.UserID, note.ResourceType, note.ResourceID, note.Title, note.Description, note.Body})
		})

		Convey("When the resource pairing CHECK fails", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23514", ConstraintName: notesResourcePairingCheckConstraint}
			}}

			_, err := repo.CreateUserNotes(context.Background(), note)

			So(errors.Is(err, notesdomain.ErrResourceDataMisMatch), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}

			_, err := repo.CreateUserNotes(context.Background(), note)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.CreateUserNotes")
		})
	})
}

func TestGetUserNotesByIDForUpdate(t *testing.T) {
	Convey("Given a notes repository", t, func() {
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, sql string, _ ...any) pgx.Row {
				gotQuery = sql
				return &testutil.MockRow{ScanFn: fakeNoteScan(fakeNote(1))}
			},
		})

		Convey("When the note is fetched for update, the row is locked", func() {
			_, err := repo.GetUserNotesByIDForUpdate(context.Background(), "note-1", "user-1")

			So(err, ShouldBeNil)
			So(gotQuery, ShouldContainSubstring, "FOR UPDATE")
			So(gotQuery, ShouldContainSubstring, "deleted_at IS NULL")
		})

		Convey("When no row matches, it is ErrNoteNotFound", func() {
			repo := newTestRepo(&testutil.MockQueryRunner{
				QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
					return &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
				},
			})

			_, err := repo.GetUserNotesByIDForUpdate(context.Background(), "note-1", "user-1")

			So(errors.Is(err, notesdomain.ErrNoteNotFound), ShouldBeTrue)
		})
	})
}

func TestGetUserNotesByID(t *testing.T) {
	Convey("Given a notes repository", t, func() {
		var row *testutil.MockRow
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, sql string, args ...any) pgx.Row {
				gotQuery, gotArgs = sql, args
				return row
			},
		})

		Convey("When the note exists, it is scoped to its owner and not soft-deleted", func() {
			note := fakeNote(1)
			row = &testutil.MockRow{ScanFn: fakeNoteScan(note)}

			got, err := repo.GetUserNotesByID(context.Background(), "note-1", "user-1")

			So(err, ShouldBeNil)
			So(got, ShouldResemble, note)
			So(gotArgs, ShouldResemble, []any{"note-1", "user-1"})
			So(gotQuery, ShouldContainSubstring, "user_id = $2")
			So(gotQuery, ShouldContainSubstring, "deleted_at IS NULL")
		})

		Convey("When no row matches, it is ErrNoteNotFound", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}

			_, err := repo.GetUserNotesByID(context.Background(), "note-1", "user-1")

			So(errors.Is(err, notesdomain.ErrNoteNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}

			_, err := repo.GetUserNotesByID(context.Background(), "note-1", "user-1")

			testutil.AssertUnexpectedDBError(err, "repository.GetUserNotesByID")
		})
	})
}

func TestGetUserAllNotesByUserID(t *testing.T) {
	testutil.TestListMethod(t, "GetUserAllNotesByUserID",
		func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*notesdomain.UserNotes, error) {
			runner.QueryRowFn = countRowFn(2, nil)
			repo := newTestRepo(runner)
			return func(ctx context.Context, params pagination.Params) ([]*notesdomain.UserNotes, error) {
				list, _, err := repo.GetUserAllNotesByUserID(ctx, "user-1", "", params)
				return list, err
			}
		}, fakeNote, fakeNoteScan)
}

func countRowFn(total int, scanErr error) func(context.Context, string, ...any) pgx.Row {
	return func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &testutil.MockRow{ScanFn: func(dest ...any) error {
			*testutil.CastInt(dest[0], 0) = total
			return scanErr
		}}
	}
}

func TestGetUserAllNotesByUserIDCount(t *testing.T) {
	Convey("Given a notes repository", t, func() {
		var countQuery string
		var countArgs []any
		var scanErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
				countQuery, countArgs = sql, args
				return countRowFn(9, scanErr)(ctx, sql, args...)
			},
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return &testutil.MockRows{}, nil
			},
		})

		Convey("When listing, the total comes from a count with the same owner and escaped search filter", func() {
			_, total, err := repo.GetUserAllNotesByUserID(context.Background(), "user-1", `50%`, pagination.NewParams(1, 10))

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 9)
			So(countQuery, ShouldContainSubstring, "COUNT(*)")
			So(countQuery, ShouldContainSubstring, "deleted_at IS NULL")
			So(countArgs, ShouldResemble, []any{"user-1", `50\%`})
		})

		Convey("When the count fails, the error mentions count and no list is returned", func() {
			scanErr = testutil.ErrDBUnexpected

			list, total, err := repo.GetUserAllNotesByUserID(context.Background(), "user-1", "", pagination.NewParams(1, 10))

			testutil.AssertUnexpectedDBError(err, "repository.GetUserAllNotesByUserID count")
			So(list, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func TestGetUserAllNotesByUserIDArgs(t *testing.T) {
	Convey("Given a notes repository", t, func() {
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
				gotQuery, gotArgs = sql, args
				return &testutil.MockRows{}, nil
			},
			QueryRowFn: countRowFn(0, nil),
		})

		Convey("When listing, the query is scoped to the owner and newest-first, with the page limit and offset", func() {
			_, _, err := repo.GetUserAllNotesByUserID(context.Background(), "user-1", "", pagination.NewParams(2, 5))

			So(err, ShouldBeNil)
			So(gotQuery, ShouldContainSubstring, "user_id = $1")
			So(gotQuery, ShouldContainSubstring, "deleted_at IS NULL")
			So(gotQuery, ShouldContainSubstring, "ORDER BY created_at DESC")
			So(gotArgs, ShouldResemble, []any{"user-1", "", 5, 5})
		})

		Convey("When searching, LIKE wildcards and backslashes in the term are escaped so they match literally", func() {
			_, _, err := repo.GetUserAllNotesByUserID(context.Background(), "user-1", `50%_off\`, pagination.NewParams(1, 10))

			So(err, ShouldBeNil)
			So(gotArgs[1], ShouldEqual, `50\%\_off\\`)
		})
	})
}

func TestUpdateUserNotes(t *testing.T) {
	Convey("Given a notes repository", t, func() {
		var tag pgconn.CommandTag
		var execErr error
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
				gotArgs = args
				return tag, execErr
			},
		})
		note := fakeNote(1)

		Convey("When it succeeds, the owner is part of the WHERE arguments", func() {
			tag = pgconn.NewCommandTag("UPDATE 1")

			So(repo.UpdateUserNotes(context.Background(), note), ShouldBeNil)
			So(gotArgs, ShouldResemble, []any{note.Title, note.Description, note.Body, note.ID, note.UserID})
		})

		Convey("When no row matches (not found or not the owner)", func() {
			tag = pgconn.NewCommandTag("UPDATE 0")

			So(errors.Is(repo.UpdateUserNotes(context.Background(), note), notesdomain.ErrNoteNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected

			testutil.AssertUnexpectedDBError(repo.UpdateUserNotes(context.Background(), note), "repository.UpdateUserNotes")
		})
	})
}

func TestDeleteUserNotes(t *testing.T) {
	testutil.TestExecMethod(t, "DeleteUserNotes",
		func(runner *testutil.MockQueryRunner) func(ctx context.Context, id, userID string) error {
			return newTestRepo(runner).DeleteUserNotes
		}, notesdomain.ErrNoteNotFound)
}
