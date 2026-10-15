package reviewservice

import (
	"context"
	"fmt"
)

// GetCourseReviewStats returns the average rating and review count for a course.
func (s *Service) GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error) {
	rating, count, err = s.courseRepo.GetCourseReviewStats(ctx, courseID)
	if err != nil {
		return 0, 0, fmt.Errorf("service.GetCourseReviewStats: %w", err)
	}
	return rating, count, nil
}

// GetContentReviewStats returns the average rating and review count for a content item.
func (s *Service) GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error) {
	rating, count, err = s.contentRepo.GetContentReviewStats(ctx, contentID)
	if err != nil {
		return 0, 0, fmt.Errorf("service.GetContentReviewStats: %w", err)
	}
	return rating, count, nil
}

// GetArticleReviewStats returns the average rating and review count for an article.
func (s *Service) GetArticleReviewStats(ctx context.Context, articleID string) (rating float64, count int, err error) {
	rating, count, err = s.articleRepo.GetArticleReviewStats(ctx, articleID)
	if err != nil {
		return 0, 0, fmt.Errorf("service.GetArticleReviewStats: %w", err)
	}
	return rating, count, nil
}
