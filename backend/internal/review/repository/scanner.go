package reviewrepository

import (
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanCourseReview(row repository.RowScanner) (*reviewdomain.CourseReview, error) {
	course := &reviewdomain.CourseReview{}
	err := row.Scan(
		&course.ID,
		&course.CourseID,
		&course.UserID,
		&course.Rating,
		&course.Comment,
		&course.CreatedAt,
		&course.UpdatedAt,
		&course.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func scanContentReview(row repository.RowScanner) (*reviewdomain.ContentReview, error) {
	content := &reviewdomain.ContentReview{}
	err := row.Scan(
		&content.ID,
		&content.ContentID,
		&content.UserID,
		&content.Rating,
		&content.Comment,
		&content.CreatedAt,
		&content.UpdatedAt,
		&content.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return content, nil
}
