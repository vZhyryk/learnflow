package reviewservice

import (
	"context"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetCourseReviews returns a paginated, optionally filtered list of reviews for a course.
func (s *Service) GetCourseReviews(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
	reviews, err := s.courseRepo.GetCourseReviewList(ctx, params, courseID, filter)
	if err != nil {
		return nil, fmt.Errorf("service.GetCourseReviews: %w", err)
	}
	return reviews, nil
}

// GetContentReviews returns a paginated, optionally filtered list of reviews for a content item.
func (s *Service) GetContentReviews(ctx context.Context, params pagination.Params, contentID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
	reviews, err := s.contentRepo.GetContentReviewList(ctx, params, contentID, filter)
	if err != nil {
		return nil, fmt.Errorf("service.GetContentReviews: %w", err)
	}
	return reviews, nil
}

// GetArticleReviews returns a paginated, optionally filtered list of reviews for an article.
func (s *Service) GetArticleReviews(ctx context.Context, params pagination.Params, articleID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error) {
	reviews, err := s.articleRepo.GetArticleReviewList(ctx, params, articleID, filter)
	if err != nil {
		return nil, fmt.Errorf("service.GetArticleReviews: %w", err)
	}
	return reviews, nil
}
