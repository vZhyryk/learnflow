package contentrepository

import (
	"context"
	"errors"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRepository(t *testing.T) {
	Convey("Given a nil connection pool", t, func() {
		Convey("NewRepository returns a non-nil Repository", func() {
			repo := NewRepository(nil)
			So(repo, ShouldNotBeNil)
		})
	})
}

func TestCreateContentItem(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a content repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeContentItem(now)
			row = &testutil.MockRow{ScanFn: fakeContentItemScan(expected)}
			got, err := repo.CreateContentItem(context.Background(), &contentdomain.ContentItem{
				Slug: expected.Slug, Title: expected.Title, CreatedByUserID: expected.CreatedByUserID,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the slug is already taken (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: "content_items_slug_unique"}
			}}
			_, err := repo.CreateContentItem(context.Background(), &contentdomain.ContentItem{})
			So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			}}
			_, err := repo.CreateContentItem(context.Background(), &contentdomain.ContentItem{})
			So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateContentItem(context.Background(), &contentdomain.ContentItem{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetContentItemByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a content repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the content item exists", func() {
			expected := fakeContentItem(now)
			row = &testutil.MockRow{ScanFn: fakeContentItemScan(expected)}
			got, err := repo.GetContentItemByID(context.Background(), "content-123")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the content item does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetContentItemByID(context.Background(), "unknown")
			So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetContentItemByID(context.Background(), "content-123")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetContentItemBySlug(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a content repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the content item exists", func() {
			expected := fakeContentItem(now)
			row = &testutil.MockRow{ScanFn: fakeContentItemScan(expected)}
			got, err := repo.GetContentItemBySlug(context.Background(), "some-slug")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the content item does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetContentItemBySlug(context.Background(), "unknown")
			So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetContentItemBySlug(context.Background(), "some-slug")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

// testExecContentItemMethod covers the shared shape of PublishContentItem/ArchiveContentItem/
// DeleteContentItem: Exec, map 0 rows affected to ErrContentItemNotFound, wrap any other error.
// Shared here instead of writing the same three Convey blocks three times over.
func testExecContentItemMethod(t *testing.T, methodName string, call func(*Repository, context.Context, string) error) {
	Convey("Given a content repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})

		Convey("When it succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(call(repo, context.Background(), "content-123"), ShouldBeNil)
		})

		Convey("When no row is matched (content item not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := call(repo, context.Background(), "unknown")
			So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := call(repo, context.Background(), "content-123")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository."+methodName)
		})
	})
}

func TestPublishContentItem(t *testing.T) {
	testExecContentItemMethod(t, "PublishContentItem", (*Repository).PublishContentItem)
}

func TestArchiveContentItem(t *testing.T) {
	testExecContentItemMethod(t, "ArchiveContentItem", (*Repository).ArchiveContentItem)
}

func TestDeleteContentItem(t *testing.T) {
	testExecContentItemMethod(t, "DeleteContentItem", (*Repository).DeleteContentItem)
}

func TestUpdateContentItem(t *testing.T) {
	Convey("Given a content repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		item := &contentdomain.ContentItem{ID: "content-123"}

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateContentItem(context.Background(), item), ShouldBeNil)
		})

		Convey("When no row is matched (content item not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateContentItem(context.Background(), item)
			So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
		})

		Convey("When the new slug is already taken (pg 23505)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "content_items_slug_unique"}
			err := repo.UpdateContentItem(context.Background(), item)
			So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			err := repo.UpdateContentItem(context.Background(), item)
			So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateContentItem(context.Background(), item)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.UpdateContentItem")
		})
	})
}

// bindContentItemList adapts a repository list method into the shape testutil.TestListMethod
// expects: build a repo around the given runner, return the bound method.
func bindContentItemList(
	call func(*Repository, context.Context, pagination.Params) ([]*contentdomain.ContentItem, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
			return call(repo, ctx, params)
		}
	}
}

// fakeContentItemN builds the nth fake ContentItem for TestListMethod's "2 items" case.
func fakeContentItemN(n int) *contentdomain.ContentItem {
	item := fakeContentItem(time.Now().UTC().Truncate(time.Second))
	item.ID = fmt.Sprintf("content-%d", n)
	return item
}

func TestGetAllPublishedContentItems(t *testing.T) {
	testutil.TestListMethod(t, "GetAllPublishedContentItems",
		bindContentItemList((*Repository).GetAllPublishedContentItems), fakeContentItemN, fakeContentItemScan)
}

func TestGetAllDraftContentItems(t *testing.T) {
	testutil.TestListMethod(t, "GetAllDraftContentItems",
		bindContentItemList((*Repository).GetAllDraftContentItems), fakeContentItemN, fakeContentItemScan)
}

func TestGetAllArchivedContentItems(t *testing.T) {
	testutil.TestListMethod(t, "GetAllArchivedContentItems",
		bindContentItemList((*Repository).GetAllArchivedContentItems), fakeContentItemN, fakeContentItemScan)
}

func TestGetAllContentItems(t *testing.T) {
	testutil.TestListMethod(t, "GetAllContentItems",
		bindContentItemList((*Repository).GetAllContentItems), fakeContentItemN, fakeContentItemScan)
}
