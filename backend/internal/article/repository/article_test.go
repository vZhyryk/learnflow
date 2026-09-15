package articlerepository

import (
	"context"
	"errors"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
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

func TestCreateArticle(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a article repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeArticle(now)
			row = &testutil.MockRow{ScanFn: fakeArticleScan(expected)}
			got, err := repo.CreateArticle(context.Background(), &articledomain.Article{
				Slug: expected.Slug, Title: expected.Title, CreatedByUserID: expected.CreatedByUserID,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the slug is already taken (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: "articles_slug_unique"}
			}}
			_, err := repo.CreateArticle(context.Background(), &articledomain.Article{})
			So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			}}
			_, err := repo.CreateArticle(context.Background(), &articledomain.Article{})
			So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateArticle(context.Background(), &articledomain.Article{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetArticleByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a article repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the article exists", func() {
			expected := fakeArticle(now)
			row = &testutil.MockRow{ScanFn: fakeArticleScan(expected)}
			got, err := repo.GetArticleByID(context.Background(), "article-123")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the article does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetArticleByID(context.Background(), "unknown")
			So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetArticleByID(context.Background(), "article-123")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetArticleBySlug(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a article repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the article exists", func() {
			expected := fakeArticle(now)
			row = &testutil.MockRow{ScanFn: fakeArticleScan(expected)}
			got, err := repo.GetArticleBySlug(context.Background(), "some-slug")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the article does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetArticleBySlug(context.Background(), "unknown")
			So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetArticleBySlug(context.Background(), "some-slug")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

// bindArticleExec adapts a repository exec method into the shape testutil.TestExecMethod
// expects: build a repo around the given runner, return the bound method.
func bindArticleExec(
	call func(*Repository, context.Context, string, string) error,
) func(*testutil.MockQueryRunner) func(context.Context, string, string) error {
	return func(runner *testutil.MockQueryRunner) func(context.Context, string, string) error {
		repo := newTestRepo(runner)
		return func(ctx context.Context, id, userID string) error {
			return call(repo, ctx, id, userID)
		}
	}
}

func TestPublishArticle(t *testing.T) {
	testutil.TestExecMethod(t, "PublishArticle", bindArticleExec((*Repository).PublishArticle), articledomain.ErrArticleNotFound)
}

func TestArchiveArticle(t *testing.T) {
	testutil.TestExecMethod(t, "ArchiveArticle", bindArticleExec((*Repository).ArchiveArticle), articledomain.ErrArticleNotFound)
}

func TestDeleteArticle(t *testing.T) {
	testutil.TestExecMethod(t, "DeleteArticle", bindArticleExec((*Repository).DeleteArticle), articledomain.ErrArticleNotFound)
}

func TestUpdateArticle(t *testing.T) {
	Convey("Given a article repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		item := &articledomain.Article{ID: "article-123"}

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateArticle(context.Background(), item, "user-1"), ShouldBeNil)
		})

		Convey("When no row is matched (article not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateArticle(context.Background(), item, "user-1")
			So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
		})

		Convey("When the new slug is already taken (pg 23505)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "articles_slug_unique"}
			err := repo.UpdateArticle(context.Background(), item, "user-1")
			So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			err := repo.UpdateArticle(context.Background(), item, "user-1")
			So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateArticle(context.Background(), item, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.UpdateArticle")
		})
	})
}

// bindArticleList adapts a repository list method into the shape testutil.TestListMethod
// expects: build a repo around the given runner, return the bound method.
func bindArticleList(
	call func(*Repository, context.Context, pagination.Params) ([]*articledomain.Article, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*articledomain.Article, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*articledomain.Article, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
			return call(repo, ctx, params)
		}
	}
}

// fakeArticleN builds the nth fake Article for TestListMethod's "2 items" case.
func fakeArticleN(n int) *articledomain.Article {
	item := fakeArticle(time.Now().UTC().Truncate(time.Second))
	item.ID = fmt.Sprintf("article-%d", n)
	return item
}

func TestGetAllPublishedArticles(t *testing.T) {
	testutil.TestListMethod(t, "GetAllPublishedArticles",
		bindArticleList((*Repository).GetAllPublishedArticles), fakeArticleN, fakeArticleScan)
}

func TestGetAllDraftArticles(t *testing.T) {
	testutil.TestListMethod(t, "GetAllDraftArticles",
		bindArticleList((*Repository).GetAllDraftArticles), fakeArticleN, fakeArticleScan)
}

func TestGetAllArchivedArticles(t *testing.T) {
	testutil.TestListMethod(t, "GetAllArchivedArticles",
		bindArticleList((*Repository).GetAllArchivedArticles), fakeArticleN, fakeArticleScan)
}

func TestGetAllArticles(t *testing.T) {
	testutil.TestListMethod(t, "GetAllArticles",
		bindArticleList((*Repository).GetAllArticles), fakeArticleN, fakeArticleScan)
}
