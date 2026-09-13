package reviewservice

import (
	"context"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
)

func (s *Service) GetCourseReviews(ctx context.Context, params pagination.Params, courseID string) ([]*reviewdomain.CourseReview, error) {
	reviews, err := s.courseRepo.GetCourseReviewList(ctx, params, courseID)
	if err != nil {
		return nil, fmt.Errorf("service.GetCourseReviews: %w", err)
	}
	return reviews, nil
}

func (s *Service) GetContentReviews(ctx context.Context, params pagination.Params, contentID string) ([]*reviewdomain.ContentReview, error) {
	reviews, err := s.contentRepo.GetContentReviewList(ctx, params, contentID)
	if err != nil {
		return nil, fmt.Errorf("service.GetContentReviews: %w", err)
	}
	return reviews, nil
}
