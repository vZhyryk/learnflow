//go:build integration

package contentrepository

import (
	"context"
	"errors"
	"testing"

	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

// draftContentItem returns a ContentItem seed with every field populated, ready for
// CreateContentItem — mirrors the shape a real CreateContentItemRequest would produce after
// Apply/validation.
func draftContentItem(t *testing.T, tx pgx.Tx) *contentdomain.ContentItem {
	t.Helper()

	description := "A thorough introduction to the subject."
	videoURL := "https://example.com/video.mp4"
	thumbnailURL := "https://example.com/thumb.png"
	estimatedMinutes := 90
	seoTitle := "SEO Title"
	seoDescription := "SEO Description"
	ogImageURL := "https://example.com/og.png"
	canonicalURL := "https://example.com/content/slug"

	return &contentdomain.ContentItem{
		Slug:             testutil.RandomTestSlug(t, "content-repo-integration"),
		Title:            "Integration Test Content Item",
		ContentType:      contentdomain.VideoContent,
		Description:      &description,
		VideoURL:         &videoURL,
		ThumbnailURL:     &thumbnailURL,
		EstimatedMinutes: &estimatedMinutes,
		SeoTitle:         &seoTitle,
		SeoDescription:   &seoDescription,
		OgImageURL:       &ogImageURL,
		CanonicalURL:     &canonicalURL,
		IsIndexable:      true,
		CreatedByUserID:  testutil.InsertRandomTestUser(t, tx),
	}
}

func TestCreateContentItem_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When creating a content item with all fields populated", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)

				got, err := repo.CreateContentItem(ctx, seed)

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.Slug, ShouldEqual, seed.Slug)
				So(got.Status, ShouldEqual, contentdomain.DraftStatus)
				So(got.IsIndexable, ShouldBeTrue)
				So(got.CreatedByUserID, ShouldEqual, seed.CreatedByUserID)
				So(got.CreatedAt.IsZero(), ShouldBeFalse)
				So(got.PublishedAt, ShouldBeNil)
				So(got.DeletedAt, ShouldBeNil)
			})
		})

		Convey("When the slug is already taken by another content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)
				_, err := repo.CreateContentItem(ctx, seed)
				So(err, ShouldBeNil)

				dup := draftContentItem(t, tx)
				dup.Slug = seed.Slug

				_, err = repo.CreateContentItem(ctx, dup)

				So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeTrue)
			})
		})

		Convey("When created_by_user_id does not reference an existing user", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)
				seed.CreatedByUserID = "00000000-0000-0000-0000-000000000000"

				_, err := repo.CreateContentItem(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503") // foreign_key_violation
			})
		})

		Convey("When title is blank (DB-level CHECK, defense-in-depth below domain validation)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)
				seed.Title = "   "

				_, err := repo.CreateContentItem(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514") // check_violation
				So(pgErr.ConstraintName, ShouldEqual, "content_items_title_nonempty")
			})
		})
	})
}

func TestGetContentItemByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When the content item exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)
				created, err := repo.CreateContentItem(ctx, seed)
				So(err, ShouldBeNil)

				got, err := repo.GetContentItemByID(ctx, created.ID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
				So(got.Slug, ShouldEqual, created.Slug)
			})
		})

		Convey("When no content item exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				_, err := repo.GetContentItemByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})

		Convey("When the content item is soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)
				created, err := repo.CreateContentItem(ctx, seed)
				So(err, ShouldBeNil)
				So(repo.DeleteContentItem(ctx, created.ID, created.CreatedByUserID), ShouldBeNil)

				_, err = repo.GetContentItemByID(ctx, created.ID)

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetContentItemBySlug_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When the content item exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftContentItem(t, tx)
				created, err := repo.CreateContentItem(ctx, seed)
				So(err, ShouldBeNil)

				got, err := repo.GetContentItemBySlug(ctx, created.Slug)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When no content item exists for the given slug", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				_, err := repo.GetContentItemBySlug(ctx, "does-not-exist")

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestPublishContentItem_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When publishing an existing draft content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)

				So(repo.PublishContentItem(ctx, created.ID, created.CreatedByUserID), ShouldBeNil)

				got, err := repo.GetContentItemByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, contentdomain.PublishedStatus)
				So(got.PublishedAt, ShouldNotBeNil)
			})
		})

		Convey("When the content item does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				err := repo.PublishContentItem(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestArchiveContentItem_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When archiving an existing draft content item directly (no publish step)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)

				So(repo.ArchiveContentItem(ctx, created.ID, created.CreatedByUserID), ShouldBeNil)

				got, err := repo.GetContentItemByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, contentdomain.ArchivedStatus)
			})
		})

		Convey("When the content item does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				err := repo.ArchiveContentItem(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteContentItem_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When soft-deleting an existing content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)

				So(repo.DeleteContentItem(ctx, created.ID, created.CreatedByUserID), ShouldBeNil)

				_, err = repo.GetContentItemByID(ctx, created.ID)
				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})

		Convey("When the content item is already deleted (second delete affects 0 rows)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)
				So(repo.DeleteContentItem(ctx, created.ID, created.CreatedByUserID), ShouldBeNil)

				err = repo.DeleteContentItem(ctx, created.ID, created.CreatedByUserID)

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateContentItem_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content repository backed by real Postgres", t, func() {
		Convey("When updating every field of an existing content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)

				newDescription := "Updated description"
				created.Slug = testutil.RandomTestSlug(t, "content-repo-integration")
				created.Title = "Updated Title"
				created.Description = &newDescription

				err = repo.UpdateContentItem(ctx, created, created.CreatedByUserID)
				So(err, ShouldBeNil)

				got, err := repo.GetContentItemByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Slug, ShouldEqual, created.Slug)
				So(got.Title, ShouldEqual, "Updated Title")
				So(*got.Description, ShouldEqual, newDescription)
			})
		})

		Convey("When the new slug collides with another content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				other, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)
				created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
				So(err, ShouldBeNil)

				created.Slug = other.Slug

				err = repo.UpdateContentItem(ctx, created, created.CreatedByUserID)

				So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeTrue)
			})
		})

		Convey("When the content item does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				ghost := draftContentItem(t, tx)
				ghost.ID = "00000000-0000-0000-0000-000000000000"

				err := repo.UpdateContentItem(ctx, ghost, ghost.CreatedByUserID)

				So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetAllContentItemsByStatus_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given draft, published, archived, and soft-deleted content items", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{repository.BaseRepository{DB: tx}}

			draft, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
			So(err, ShouldBeNil)

			published, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
			So(err, ShouldBeNil)
			So(repo.PublishContentItem(ctx, published.ID, published.CreatedByUserID), ShouldBeNil)

			archived, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
			So(err, ShouldBeNil)
			So(repo.ArchiveContentItem(ctx, archived.ID, archived.CreatedByUserID), ShouldBeNil)

			deleted, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
			So(err, ShouldBeNil)
			So(repo.DeleteContentItem(ctx, deleted.ID, deleted.CreatedByUserID), ShouldBeNil)

			params := pagination.NewParams(1, 100)

			Convey("GetAllDraftContentItems returns only the draft, excluding soft-deleted", func() {
				got, err := repo.GetAllDraftContentItems(ctx, params)
				So(err, ShouldBeNil)
				ids := contentItemIDs(got)
				So(ids, ShouldContain, draft.ID)
				So(ids, ShouldNotContain, deleted.ID)
			})

			Convey("GetAllPublishedContentItems returns only the published content item", func() {
				got, err := repo.GetAllPublishedContentItems(ctx, params)
				So(err, ShouldBeNil)
				ids := contentItemIDs(got)
				So(ids, ShouldContain, published.ID)
				So(ids, ShouldNotContain, draft.ID)
				So(ids, ShouldNotContain, archived.ID)
			})

			Convey("GetAllArchivedContentItems includes the soft-deleted archived content item too", func() {
				// GetAllArchivedContentItems intentionally omits `deleted_at IS NULL` — it's an
				// admin "including soft-deleted ones" query per db-conventions.md.
				So(repo.DeleteContentItem(ctx, archived.ID, archived.CreatedByUserID), ShouldBeNil)

				got, err := repo.GetAllArchivedContentItems(ctx, params)
				So(err, ShouldBeNil)
				So(contentItemIDs(got), ShouldContain, archived.ID)
			})

			Convey("GetAllContentItems returns every content item regardless of status, including soft-deleted", func() {
				got, err := repo.GetAllContentItems(ctx, params)
				So(err, ShouldBeNil)
				ids := contentItemIDs(got)
				So(ids, ShouldContain, draft.ID)
				So(ids, ShouldContain, published.ID)
				So(ids, ShouldContain, archived.ID)
				So(ids, ShouldContain, deleted.ID)
			})

			Convey("Pagination limits the returned page size", func() {
				got, err := repo.GetAllContentItems(ctx, pagination.NewParams(1, 2))
				So(err, ShouldBeNil)
				So(len(got), ShouldEqual, 2)
			})
		})
	})
}

func contentItemIDs(items []*contentdomain.ContentItem) []string {
	ids := make([]string, 0, len(items))
	for _, c := range items {
		ids = append(ids, c.ID)
	}
	return ids
}

func TestContentItemCreatedByUserForeignKeyRestrict_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content item created by an existing user", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{repository.BaseRepository{DB: tx}}
			created, err := repo.CreateContentItem(ctx, draftContentItem(t, tx))
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
