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
		&course.DeletedByUserID,
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
		&content.DeletedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func scanArticleReview(row repository.RowScanner) (*reviewdomain.ArticleReview, error) {
	article := &reviewdomain.ArticleReview{}
	err := row.Scan(
		&article.ID,
		&article.ArticleID,
		&article.UserID,
		&article.Rating,
		&article.Comment,
		&article.CreatedAt,
		&article.UpdatedAt,
		&article.DeletedAt,
		&article.DeletedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return article, nil
}
