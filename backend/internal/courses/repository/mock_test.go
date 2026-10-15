package courserepository

import (
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeCourse(now time.Time) *coursedomain.Course {
	description := "description"
	thumbnailURL := "https://example.com/thumb.png"
	previewVideoURL := "https://example.com/preview.mp4"
	seoTitle := "seo title"
	seoDescription := "seo description"
	ogImageURL := "https://example.com/og.png"
	canonicalURL := "https://example.com/course"
	estimatedMinutes := 42
	publishedAt := now
	updatedByUserID := "user-456"
	publishedByUserID := "user-456"

	return &coursedomain.Course{
		ID:                "course-123",
		Slug:              "some-slug",
		Title:             "Some Title",
		Description:       &description,
		ThumbnailURL:      &thumbnailURL,
		PreviewVideoURL:   &previewVideoURL,
		Status:            coursedomain.PublishedStatus,
		EstimatedMinutes:  &estimatedMinutes,
		SeoTitle:          &seoTitle,
		SeoDescription:    &seoDescription,
		OgImageURL:        &ogImageURL,
		CanonicalURL:      &canonicalURL,
		IsIndexable:       true,
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

// fakeCourseScan simulates rows.Scan populating a Course from column order, matching
// scanCourse in scanner.go. Reused across CreateCourse, GetCourseByID, GetCourseBySlug,
// and every GetAll*Courses method.
func fakeCourseScan(course *coursedomain.Course) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = course.ID
		*testutil.CastStr(dest[1], 1) = course.Slug
		*testutil.CastStr(dest[2], 2) = course.Title
		*testutil.CastPtrStr(dest[3], 3) = course.Description
		*testutil.CastPtrStr(dest[4], 4) = course.ThumbnailURL
		*testutil.CastPtrStr(dest[5], 5) = course.PreviewVideoURL
		*testutil.CastEnum[coursedomain.CourseStatus](dest[6], 6) = course.Status
		*testutil.CastPtrInt(dest[7], 7) = course.EstimatedMinutes
		*testutil.CastPtrStr(dest[8], 8) = course.SeoTitle
		*testutil.CastPtrStr(dest[9], 9) = course.SeoDescription
		*testutil.CastPtrStr(dest[10], 10) = course.OgImageURL
		*testutil.CastPtrStr(dest[11], 11) = course.CanonicalURL
		*testutil.CastBool(dest[12], 12) = course.IsIndexable
		*testutil.CastStr(dest[13], 13) = course.CreatedByUserID
		*testutil.CastTime(dest[14], 14) = course.CreatedAt
		*testutil.CastPtrTime(dest[15], 15) = course.UpdatedAt
		*testutil.CastPtrStr(dest[16], 16) = course.UpdatedByUserID
		*testutil.CastPtrTime(dest[17], 17) = course.PublishedAt
		*testutil.CastPtrStr(dest[18], 18) = course.PublishedByUserID
		*testutil.CastPtrTime(dest[19], 19) = course.DeletedAt
		*testutil.CastPtrStr(dest[20], 20) = course.DeletedByUserID
		*testutil.CastPtrTime(dest[21], 21) = course.ArchivedAt
		*testutil.CastPtrStr(dest[22], 22) = course.ArchivedByUserID
		return nil
	}
}
