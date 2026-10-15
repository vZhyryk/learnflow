package contentrepository

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeContentItem(now time.Time) *contentdomain.ContentItem {
	description := "description"
	videoURL := "https://example.com/video.mp4"
	estimatedMinutes := 42
	thumbnailURL := "https://example.com/thumb.png"
	seoTitle := "seo title"
	seoDescription := "seo description"
	ogImageURL := "https://example.com/og.png"
	canonicalURL := "https://example.com/content-item"
	publishedAt := now
	updatedByUserID := "user-456"
	publishedByUserID := "user-456"

	return &contentdomain.ContentItem{
		ID:                "content-item-123",
		Slug:              "some-slug",
		Title:             "Some Title",
		ContentType:       contentdomain.VideoContent,
		Description:       &description,
		VideoURL:          &videoURL,
		EstimatedMinutes:  &estimatedMinutes,
		ThumbnailURL:      &thumbnailURL,
		SeoTitle:          &seoTitle,
		SeoDescription:    &seoDescription,
		OgImageURL:        &ogImageURL,
		CanonicalURL:      &canonicalURL,
		IsIndexable:       true,
		Status:            contentdomain.PublishedStatus,
		CreatedByUserID:   "user-123",
		CreatedAt:         now,
		UpdatedAt:         &now,
		UpdatedByUserID:   &updatedByUserID,
		PublishedAt:       &publishedAt,
		PublishedByUserID: &publishedByUserID,
		DeletedAt:         nil,
		DeletedByUserID:   nil,
		ArchivedAt:        nil,
		ArchivedByUserID:  nil,
	}
}

// fakeContentItemScan simulates rows.Scan populating a ContentItem from column order, matching
// scanContentItem in scanner.go. Reused across CreateContentItem, GetContentItemByID,
// GetContentItemBySlug, and every GetAll*ContentItems method.
func fakeContentItemScan(item *contentdomain.ContentItem) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = item.ID
		*testutil.CastStr(dest[1], 1) = item.Slug
		*testutil.CastStr(dest[2], 2) = item.Title
		*testutil.CastEnum[contentdomain.ContentType](dest[3], 3) = item.ContentType
		*testutil.CastPtrStr(dest[4], 4) = item.Description
		*testutil.CastPtrStr(dest[5], 5) = item.Body
		*testutil.CastPtrStr(dest[6], 6) = item.VideoURL
		*testutil.CastPtrStr(dest[7], 7) = item.FileURL
		*testutil.CastPtrInt(dest[8], 8) = item.EstimatedMinutes
		*testutil.CastPtrInt(dest[9], 9) = item.EstimatedPages
		*testutil.CastPtrStr(dest[10], 10) = item.ThumbnailURL
		*testutil.CastPtrStr(dest[11], 11) = item.SeoTitle
		*testutil.CastPtrStr(dest[12], 12) = item.SeoDescription
		*testutil.CastPtrStr(dest[13], 13) = item.OgImageURL
		*testutil.CastPtrStr(dest[14], 14) = item.CanonicalURL
		*testutil.CastBool(dest[15], 15) = item.IsIndexable
		*testutil.CastEnum[contentdomain.ContentItemStatus](dest[16], 16) = item.Status
		*testutil.CastStr(dest[17], 17) = item.CreatedByUserID
		*testutil.CastTime(dest[18], 18) = item.CreatedAt
		*testutil.CastPtrTime(dest[19], 19) = item.UpdatedAt
		*testutil.CastPtrStr(dest[20], 20) = item.UpdatedByUserID
		*testutil.CastPtrTime(dest[21], 21) = item.PublishedAt
		*testutil.CastPtrStr(dest[22], 22) = item.PublishedByUserID
		*testutil.CastPtrTime(dest[23], 23) = item.DeletedAt
		*testutil.CastPtrStr(dest[24], 24) = item.DeletedByUserID
		*testutil.CastPtrTime(dest[25], 25) = item.ArchivedAt
		*testutil.CastPtrStr(dest[26], 26) = item.ArchivedByUserID
		return nil
	}
}
