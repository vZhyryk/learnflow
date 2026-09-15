package courserepository

import (
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanCourse(row repository.RowScanner) (*coursedomain.Course, error) {
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
		&course.UpdatedByUserID,
		&course.PublishedAt,
		&course.PublishedByUserID,
		&course.DeletedAt,
		&course.DeletedByUserID,
		&course.ArchivedAt,
		&course.ArchivedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return course, nil
}
