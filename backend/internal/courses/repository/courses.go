package courserepository

import (
	"context"
	"errors"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
)

// coursesSlugUniqueConstraint is the DB-level backstop for the check-then-insert slug race.
const coursesSlugUniqueConstraint = "courses_slug_unique"

// CreateCourse inserts a new draft course.
func (rep *Repository) CreateCourse(ctx context.Context, course *coursedomain.Course) (*coursedomain.Course, error) {
	course, err := scanCourse(rep.QueryRunner(ctx).QueryRow(ctx, createDraftCourseSQL, course.Slug, course.Title, course.Description, course.ThumbnailURL, course.PreviewVideoURL, course.EstimatedMinutes, course.SeoTitle, course.SeoDescription, course.OgImageURL, course.CanonicalURL, course.IsIndexable, course.CreatedByUserID))
	if db.IsUniqueViolation(err, coursesSlugUniqueConstraint) {
		return nil, coursedomain.ErrInvalidSlug
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateCourse: %w", err)
	}

	return course, nil
}

// PublishCourse marks a course as published.
func (rep *Repository) PublishCourse(ctx context.Context, courseID, userID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, publishCourseSQL, "PublishCourse", courseID, userID, coursedomain.ErrCourseNotFound)
}

// ArchiveCourse marks a course as archived.
func (rep *Repository) ArchiveCourse(ctx context.Context, courseID, userID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, archiveCourseSQL, "ArchiveCourse", courseID, userID, coursedomain.ErrCourseNotFound)
}

// DeleteCourse soft-deletes a course.
func (rep *Repository) DeleteCourse(ctx context.Context, courseID, userID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, deleteCourseSQL, "DeleteCourse", courseID, userID, coursedomain.ErrCourseNotFound)
}

// UpdateCourse persists changes to an existing course.
func (rep *Repository) UpdateCourse(ctx context.Context, course *coursedomain.Course, userID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateCourseSQL, course.ID, course.Slug, course.Title, course.Description, course.ThumbnailURL, course.PreviewVideoURL, course.EstimatedMinutes, course.SeoTitle, course.SeoDescription, course.OgImageURL, course.CanonicalURL, course.IsIndexable, userID)
	if db.IsUniqueViolation(err, coursesSlugUniqueConstraint) {
		return coursedomain.ErrInvalidSlug
	}
	if err != nil {
		return fmt.Errorf("repository.UpdateCourse: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return coursedomain.ErrCourseNotFound
	}

	return nil
}

// GetAllPublishedCourses returns every non-deleted published course.
func (rep *Repository) GetAllPublishedCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllPublishedCoursesSQL, "GetAllPublishedCourses", params, scanCourse)
}

// GetAllDraftCourses returns every non-deleted draft course.
func (rep *Repository) GetAllDraftCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllDraftCoursesSQL, "GetAllDraftCourses", params, scanCourse)
}

// GetAllArchivedCourses returns every archived course, including soft-deleted ones.
func (rep *Repository) GetAllArchivedCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllArchivedCoursesSQL, "GetAllArchivedCourses", params, scanCourse)
}

// GetAllCourses returns every course regardless of status, including soft-deleted ones.
func (rep *Repository) GetAllCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllCoursesSQL, "GetAllCourses", params, scanCourse)
}

// GetCourseByID retrieves a non-deleted course by ID.
func (rep *Repository) GetCourseByID(ctx context.Context, courseID string) (*coursedomain.Course, error) {
	course, err := scanCourse(rep.QueryRunner(ctx).QueryRow(ctx, getCourseByIDSQL, courseID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, coursedomain.ErrCourseNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetCourseByID: %w", err)
	}

	return course, nil
}

// GetCourseBySlug retrieves a non-deleted course by slug.
func (rep *Repository) GetCourseBySlug(ctx context.Context, slug string) (*coursedomain.Course, error) {
	course, err := scanCourse(rep.QueryRunner(ctx).QueryRow(ctx, getCourseBySlugSQL, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, coursedomain.ErrCourseNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetCourseBySlug: %w", err)
	}

	return course, nil
}
