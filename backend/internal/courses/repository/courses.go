package courserepository

import (
	"context"
	"errors"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// slugUniqueViolation reports whether err is a Postgres unique_violation (23505) on
// courses_slug_unique — the DB-level backstop for the check-then-insert slug race.
func slugUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "courses_slug_unique"
}

// CreateCourse inserts a new draft course.
func (rep *Repository) CreateCourse(ctx context.Context, course *coursedomain.Course) (*coursedomain.Course, error) {
	course, err := scanCourse(rep.queryRunner(ctx).QueryRow(ctx, createDraftCourseSQL, course.Slug, course.Title, course.Description, course.ThumbnailURL, course.PreviewVideoURL, course.EstimatedMinutes, course.SeoTitle, course.SeoDescription, course.OgImageURL, course.CanonicalURL, course.IsIndexable, course.CreatedByUserID))
	if slugUniqueViolation(err) {
		return nil, coursedomain.ErrInvalidSlug
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateCourse: %w", err)
	}

	return course, nil
}

// Below: single UPDATE ... WHERE id = $1 statements — no prior SELECT needed, unlike
// auth/repository's FOR UPDATE NOWAIT flow which reads then conditionally writes.

// execCourseStatusChange runs a status-change UPDATE, wrapping errors and mapping 0 rows
// affected to ErrCourseNotFound.
func (rep *Repository) execCourseStatusChange(ctx context.Context, sql, methodName, courseID string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, sql, courseID)
	if err != nil {
		return fmt.Errorf("repository.%s: %w", methodName, err)
	}

	if tag.RowsAffected() == 0 {
		return coursedomain.ErrCourseNotFound
	}

	return nil
}

// PublishCourse marks a course as published.
func (rep *Repository) PublishCourse(ctx context.Context, courseID string) error {
	return rep.execCourseStatusChange(ctx, publishCourseSQL, "PublishCourse", courseID)
}

// ArchiveCourse marks a course as archived.
func (rep *Repository) ArchiveCourse(ctx context.Context, courseID string) error {
	return rep.execCourseStatusChange(ctx, archiveCourseSQL, "ArchiveCourse", courseID)
}

// DeleteCourse soft-deletes a course.
func (rep *Repository) DeleteCourse(ctx context.Context, courseID string) error {
	return rep.execCourseStatusChange(ctx, deleteCourseSQL, "DeleteCourse", courseID)
}

// UpdateCourse persists changes to an existing course.
func (rep *Repository) UpdateCourse(ctx context.Context, course *coursedomain.Course) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateCourseSQL, course.ID, course.Slug, course.Title, course.Description, course.ThumbnailURL, course.PreviewVideoURL, course.EstimatedMinutes, course.SeoTitle, course.SeoDescription, course.OgImageURL, course.CanonicalURL, course.IsIndexable)
	if slugUniqueViolation(err) {
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
	return rep.getAndParseCourses(ctx, getAllPublishedCoursesSQL, "GetAllPublishedCourses", params)
}

// GetAllDraftCourses returns every non-deleted draft course.
func (rep *Repository) GetAllDraftCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return rep.getAndParseCourses(ctx, getAllDraftCoursesSQL, "GetAllDraftCourses", params)
}

// GetAllArchivedCourses returns every archived course, including soft-deleted ones.
func (rep *Repository) GetAllArchivedCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return rep.getAndParseCourses(ctx, getAllArchivedCoursesSQL, "GetAllArchivedCourses", params)
}

// GetAllCourses returns every course regardless of status, including soft-deleted ones.
func (rep *Repository) GetAllCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	return rep.getAndParseCourses(ctx, getAllCoursesSQL, "GetAllCourses", params)
}

// GetCourseByID retrieves a non-deleted course by ID.
func (rep *Repository) GetCourseByID(ctx context.Context, courseID string) (*coursedomain.Course, error) {
	course, err := scanCourse(rep.queryRunner(ctx).QueryRow(ctx, getCourseByIDSQL, courseID))
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
	course, err := scanCourse(rep.queryRunner(ctx).QueryRow(ctx, getCourseBySlugSQL, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, coursedomain.ErrCourseNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetCourseBySlug: %w", err)
	}

	return course, nil
}
