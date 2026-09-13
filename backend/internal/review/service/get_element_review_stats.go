package reviewservice

import (
	"context"
	"fmt"
)

func (s *Service) GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error) {
	rating, count, err = s.courseRepo.GetCourseReviewStats(ctx, courseID)
	if err != nil {
		return 0, 0, fmt.Errorf("service.GetCourseReviewStats: %w", err)
	}
	return rating, count, nil
}

func (s *Service) GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error) {
	rating, count, err = s.contentRepo.GetContentReviewStats(ctx, contentID)
	if err != nil {
		return 0, 0, fmt.Errorf("service.GetContentReviewStats: %w", err)
	}
	return rating, count, nil
}
