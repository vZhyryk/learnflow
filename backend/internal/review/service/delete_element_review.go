package reviewservice

import (
	"context"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
)

func (s *Service) DeleteCourseReview(ctx context.Context, reviewID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.courseRepo.GetCourseReviewByID(ctx, reviewID)
		if err != nil {
			return fmt.Errorf("service.DeleteCourseReview: fetch review: %w", err)
		}
		if currentReview.UserID != userID {
			return reviewdomain.ErrReviewNotFound
		}

		if err = s.courseRepo.DeleteCourseReview(ctx, reviewID, userID); err != nil {
			return fmt.Errorf("service.DeleteCourseReview: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteContentReview(ctx context.Context, reviewID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.contentRepo.GetContentReviewByID(ctx, reviewID)
		if err != nil {
			return fmt.Errorf("service.DeleteContentReview: fetch review: %w", err)
		}
		if currentReview.UserID != userID {
			return reviewdomain.ErrReviewNotFound
		}

		if err := s.contentRepo.DeleteContentReview(ctx, reviewID, userID); err != nil {
			return fmt.Errorf("service.DeleteContentReview: %w", err)
		}
		return nil
	})

}

func (s *Service) DeleteCourseReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if err := s.courseRepo.DeleteCourseReview(ctx, reviewID, userID); err != nil {
		return fmt.Errorf("service.DeleteCourseReviewAdmin: %w", err)
	}
	return nil
}

func (s *Service) DeleteContentReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if err := s.contentRepo.DeleteContentReview(ctx, reviewID, userID); err != nil {
		return fmt.Errorf("service.DeleteContentReviewAdmin: %w", err)
	}
	return nil
}
