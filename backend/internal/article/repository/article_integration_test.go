//go:build integration

package articlerepository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

func randomTestSlug(t *testing.T) string {
	t.Helper()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("randomTestSlug: %v", err)
	}
	return fmt.Sprintf("article-repo-integration-%s", hex.EncodeToString(buf))
}

func insertTestUser(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "article-repo-integration"))
}

// draftArticle returns a Article seed with every field populated, ready for
// CreateArticle — mirrors the shape a real CreateArticleRequest would produce after
// Apply/validation.
func draftArticle(t *testing.T, tx pgx.Tx) *articledomain.Article {
	t.Helper()

	seoTitle := "SEO Title"
	seoDescription := "SEO Description"
	ogImageURL := "https://example.com/og.png"

	return &articledomain.Article{
		Slug:            randomTestSlug(t),
		Title:           "Integration Test Article",
		SeoTitle:        &seoTitle,
		SeoDescription:  &seoDescription,
		OgImageURL:      &ogImageURL,
		IsIndexable:     true,
		CreatedByUserID: insertTestUser(t, tx),
	}
}

func TestCreateArticle_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When creating a article with all fields populated", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)

				got, err := repo.CreateArticle(ctx, seed)

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.Slug, ShouldEqual, seed.Slug)
				So(got.Status, ShouldEqual, articledomain.DraftStatus)
				So(got.IsIndexable, ShouldBeTrue)
				So(got.CreatedByUserID, ShouldEqual, seed.CreatedByUserID)
				So(got.CreatedAt.IsZero(), ShouldBeFalse)
				So(got.PublishedAt, ShouldBeNil)
				So(got.DeletedAt, ShouldBeNil)
			})
		})

		Convey("When the slug is already taken by another article", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)
				_, err := repo.CreateArticle(ctx, seed)
				So(err, ShouldBeNil)

				dup := draftArticle(t, tx)
				dup.Slug = seed.Slug

				_, err = repo.CreateArticle(ctx, dup)

				So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeTrue)
			})
		})

		Convey("When created_by_user_id does not reference an existing user", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)
				seed.CreatedByUserID = "00000000-0000-0000-0000-000000000000"

				_, err := repo.CreateArticle(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503") // foreign_key_violation
			})
		})

		Convey("When title is blank (DB-level CHECK, defense-in-depth below domain validation)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)
				seed.Title = "   "

				_, err := repo.CreateArticle(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514") // check_violation
				So(pgErr.ConstraintName, ShouldEqual, "articles_title_nonempty")
			})
		})
	})
}

func TestGetArticleByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When the article exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)
				created, err := repo.CreateArticle(ctx, seed)
				So(err, ShouldBeNil)

				got, err := repo.GetArticleByID(ctx, created.ID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
				So(got.Slug, ShouldEqual, created.Slug)
			})
		})

		Convey("When no article exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				_, err := repo.GetArticleByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})

		Convey("When the article is soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)
				created, err := repo.CreateArticle(ctx, seed)
				So(err, ShouldBeNil)
				So(repo.DeleteArticle(ctx, created.ID), ShouldBeNil)

				_, err = repo.GetArticleByID(ctx, created.ID)

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetArticleBySlug_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When the article exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftArticle(t, tx)
				created, err := repo.CreateArticle(ctx, seed)
				So(err, ShouldBeNil)

				got, err := repo.GetArticleBySlug(ctx, created.Slug)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When no article exists for the given slug", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				_, err := repo.GetArticleBySlug(ctx, "does-not-exist")

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestPublishArticle_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When publishing an existing draft article", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)

				So(repo.PublishArticle(ctx, created.ID), ShouldBeNil)

				got, err := repo.GetArticleByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, articledomain.PublishedStatus)
				So(got.PublishedAt, ShouldNotBeNil)
			})
		})

		Convey("When the article does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				err := repo.PublishArticle(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestArchiveArticle_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When archiving an existing draft article directly (no publish step)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)

				So(repo.ArchiveArticle(ctx, created.ID), ShouldBeNil)

				got, err := repo.GetArticleByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, articledomain.ArchivedStatus)
			})
		})

		Convey("When the article does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				err := repo.ArchiveArticle(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteArticle_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When soft-deleting an existing article", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)

				So(repo.DeleteArticle(ctx, created.ID), ShouldBeNil)

				_, err = repo.GetArticleByID(ctx, created.ID)
				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})

		Convey("When the article is already deleted (second delete affects 0 rows)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)
				So(repo.DeleteArticle(ctx, created.ID), ShouldBeNil)

				err = repo.DeleteArticle(ctx, created.ID)

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateArticle_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article repository backed by real Postgres", t, func() {
		Convey("When updating every field of an existing article", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)

				created.Slug = randomTestSlug(t)
				created.Title = "Updated Title"

				err = repo.UpdateArticle(ctx, created)
				So(err, ShouldBeNil)

				got, err := repo.GetArticleByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Slug, ShouldEqual, created.Slug)
				So(got.Title, ShouldEqual, "Updated Title")
			})
		})

		Convey("When the new slug collides with another article", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				other, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)
				created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
				So(err, ShouldBeNil)

				created.Slug = other.Slug

				err = repo.UpdateArticle(ctx, created)

				So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeTrue)
			})
		})

		Convey("When the article does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				ghost := draftArticle(t, tx)
				ghost.ID = "00000000-0000-0000-0000-000000000000"

				err := repo.UpdateArticle(ctx, ghost)

				So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetAllArticlesByStatus_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given draft, published, archived, and soft-deleted articles", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{repository.BaseRepository{DB: tx}}

			draft, err := repo.CreateArticle(ctx, draftArticle(t, tx))
			So(err, ShouldBeNil)

			published, err := repo.CreateArticle(ctx, draftArticle(t, tx))
			So(err, ShouldBeNil)
			So(repo.PublishArticle(ctx, published.ID), ShouldBeNil)

			archived, err := repo.CreateArticle(ctx, draftArticle(t, tx))
			So(err, ShouldBeNil)
			So(repo.ArchiveArticle(ctx, archived.ID), ShouldBeNil)

			deleted, err := repo.CreateArticle(ctx, draftArticle(t, tx))
			So(err, ShouldBeNil)
			So(repo.DeleteArticle(ctx, deleted.ID), ShouldBeNil)

			params := pagination.NewParams(1, 100)

			Convey("GetAllDraftArticles returns only the draft, excluding soft-deleted", func() {
				got, err := repo.GetAllDraftArticles(ctx, params)
				So(err, ShouldBeNil)
				ids := ArticleIDs(got)
				So(ids, ShouldContain, draft.ID)
				So(ids, ShouldNotContain, deleted.ID)
			})

			Convey("GetAllPublishedArticles returns only the published article", func() {
				got, err := repo.GetAllPublishedArticles(ctx, params)
				So(err, ShouldBeNil)
				ids := ArticleIDs(got)
				So(ids, ShouldContain, published.ID)
				So(ids, ShouldNotContain, draft.ID)
				So(ids, ShouldNotContain, archived.ID)
			})

			Convey("GetAllArchivedArticles includes the soft-deleted archived article too", func() {
				// GetAllArchivedArticles intentionally omits `deleted_at IS NULL` — it's an
				// admin "including soft-deleted ones" query per db-conventions.md.
				So(repo.DeleteArticle(ctx, archived.ID), ShouldBeNil)

				got, err := repo.GetAllArchivedArticles(ctx, params)
				So(err, ShouldBeNil)
				So(ArticleIDs(got), ShouldContain, archived.ID)
			})

			Convey("GetAllArticles returns every article regardless of status, including soft-deleted", func() {
				got, err := repo.GetAllArticles(ctx, params)
				So(err, ShouldBeNil)
				ids := ArticleIDs(got)
				So(ids, ShouldContain, draft.ID)
				So(ids, ShouldContain, published.ID)
				So(ids, ShouldContain, archived.ID)
				So(ids, ShouldContain, deleted.ID)
			})

			Convey("Pagination limits the returned page size", func() {
				got, err := repo.GetAllArticles(ctx, pagination.NewParams(1, 2))
				So(err, ShouldBeNil)
				So(len(got), ShouldEqual, 2)
			})
		})
	})
}

func ArticleIDs(items []*articledomain.Article) []string {
	ids := make([]string, 0, len(items))
	for _, c := range items {
		ids = append(ids, c.ID)
	}
	return ids
}

func TestArticleCreatedByUserForeignKeyRestrict_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a article created by an existing user", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{repository.BaseRepository{DB: tx}}
			created, err := repo.CreateArticle(ctx, draftArticle(t, tx))
			So(err, ShouldBeNil)

			Convey("When hard-deleting the owning user row", func() {
				_, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, created.CreatedByUserID)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503")
			})
		})
	})
}
