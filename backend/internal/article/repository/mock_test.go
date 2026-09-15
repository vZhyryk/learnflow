package articlerepository

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeArticle(now time.Time) *articledomain.Article {
	seoTitle := "seo title"
	seoDescription := "seo description"
	ogImageURL := "https://example.com/og.png"
	description := "Description"
	publishedAt := now

	return &articledomain.Article{
		ID:              "article-123",
		Slug:            "some-slug",
		Title:           "Some Title",
		Description:     &description,
		SeoTitle:        &seoTitle,
		SeoDescription:  &seoDescription,
		OgImageURL:      &ogImageURL,
		IsIndexable:     true,
		Status:          articledomain.PublishedStatus,
		CreatedByUserID: "user-123",
		CreatedAt:       now,
		UpdatedAt:       now,
		PublishedAt:     &publishedAt,
		DeletedAt:       nil,
	}
}

func fakeArticleScan(item *articledomain.Article) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = item.ID
		*testutil.CastStr(dest[1], 1) = item.Slug
		*testutil.CastStr(dest[2], 2) = item.Title
		*testutil.CastPtrStr(dest[3], 3) = item.Description
		*testutil.CastStr(dest[4], 4) = item.Body
		*testutil.CastPtrStr(dest[5], 5) = item.SeoTitle
		*testutil.CastPtrStr(dest[6], 6) = item.SeoDescription
		*testutil.CastPtrStr(dest[7], 7) = item.OgImageURL
		*testutil.CastBool(dest[8], 8) = item.IsIndexable
		*testutil.CastEnum[articledomain.ArticleStatus](dest[9], 9) = item.Status
		*testutil.CastStr(dest[10], 10) = item.CreatedByUserID
		*testutil.CastTime(dest[11], 11) = item.CreatedAt
		*testutil.CastTime(dest[12], 12) = item.UpdatedAt
		*testutil.CastPtrTime(dest[13], 13) = item.PublishedAt
		*testutil.CastPtrTime(dest[14], 14) = item.DeletedAt
		return nil
	}
}
