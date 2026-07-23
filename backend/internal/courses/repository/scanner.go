package courserepository

import (
	"context"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCourse(row rowScanner) (*coursedomain.Course, error) {
	course := &coursedomain.Course{}
	err := row.Scan(
		&course.ID,
		&course.Slug,
		&course.Title,
		&course.Description,
		&course.ThumbnailURL,
		&course.PreviewVideoURL,
		&course.Status,
		&course.EstimatedMinutes,
		&course.SeoTitle,
		&course.SeoDescription,
		&course.OgImageURL,
		&course.CanonicalURL,
		&course.IsIndexable,
		&course.CreatedByUserID,
		&course.CreatedAt,
		&course.UpdatedAt,
		&course.PublishedAt,
		&course.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (rep *Repository) getAndParseCourses(ctx context.Context, query, methodName string, params pagination.Params) ([]*coursedomain.Course, error) {
	rows, err := rep.queryRunner(ctx).Query(ctx, query, params.Limit(), params.Offset())
	if err != nil {
		return nil, fmt.Errorf("repository.%s: %w", methodName, err)
	}

	defer rows.Close()

	var courses []*coursedomain.Course
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.%s scan: %w", methodName, err)
		}
		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.%s rows: %w", methodName, err)
	}

	return courses, nil
}
