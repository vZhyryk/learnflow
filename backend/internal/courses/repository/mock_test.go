package courserepository

import (
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/testutil"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{db: runner}
}

// castCourseStatus type-asserts a scan destination to *coursedomain.CourseStatus — a
// domain-specific enum, so this stays package-local rather than in shared testutil
// (mirrors castUserRole/castUserStatus in internal/auth/repository/mock_test.go).
func castCourseStatus(v any, idx int) *coursedomain.CourseStatus {
	s, ok := v.(*coursedomain.CourseStatus)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *coursedomain.CourseStatus, got %T", idx, v))
	}
	return s
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

	return &coursedomain.Course{
		ID:               "course-123",
		Slug:             "some-slug",
		Title:            "Some Title",
		Description:      &description,
		ThumbnailURL:     &thumbnailURL,
		PreviewVideoURL:  &previewVideoURL,
		Status:           coursedomain.PublishedStatus,
		EstimatedMinutes: &estimatedMinutes,
		SeoTitle:         &seoTitle,
		SeoDescription:   &seoDescription,
		OgImageURL:       &ogImageURL,
		CanonicalURL:     &canonicalURL,
		IsIndexable:      true,
		CreatedByUserID:  "user-123",
		CreatedAt:        now,
		UpdatedAt:        now,
		PublishedAt:      &publishedAt,
		DeletedAt:        nil,
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
		*castCourseStatus(dest[6], 6) = course.Status
		*testutil.CastPtrInt(dest[7], 7) = course.EstimatedMinutes
		*testutil.CastPtrStr(dest[8], 8) = course.SeoTitle
		*testutil.CastPtrStr(dest[9], 9) = course.SeoDescription
		*testutil.CastPtrStr(dest[10], 10) = course.OgImageURL
		*testutil.CastPtrStr(dest[11], 11) = course.CanonicalURL
		*testutil.CastBool(dest[12], 12) = course.IsIndexable
		*testutil.CastStr(dest[13], 13) = course.CreatedByUserID
		*testutil.CastTime(dest[14], 14) = course.CreatedAt
		*testutil.CastTime(dest[15], 15) = course.UpdatedAt
		*testutil.CastPtrTime(dest[16], 16) = course.PublishedAt
		*testutil.CastPtrTime(dest[17], 17) = course.DeletedAt
		return nil
	}
}
