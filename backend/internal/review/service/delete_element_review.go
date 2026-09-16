package reviewservice

import (
	"context"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
)

// DeleteCourseReview soft-deletes a course review owned by the requesting user.
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

// DeleteContentReview soft-deletes a content review owned by the requesting user.
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

// DeleteCourseReviewAdmin soft-deletes any course review, bypassing ownership checks.
func (s *Service) DeleteCourseReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if err := s.courseRepo.DeleteCourseReview(ctx, reviewID, userID); err != nil {
		return fmt.Errorf("service.DeleteCourseReviewAdmin: %w", err)
	}
	return nil
}

// DeleteContentReviewAdmin soft-deletes any content review, bypassing ownership checks.
func (s *Service) DeleteContentReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if err := s.contentRepo.DeleteContentReview(ctx, reviewID, userID); err != nil {
		return fmt.Errorf("service.DeleteContentReviewAdmin: %w", err)
	}
	return nil
}

// DeleteArticleReview soft-deletes an article review owned by the requesting user.
func (s *Service) DeleteArticleReview(ctx context.Context, reviewID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.articleRepo.GetArticleReviewByID(ctx, reviewID)
		if err != nil {
			return fmt.Errorf("service.DeleteArticleReview: fetch review: %w", err)
		}
		if currentReview.UserID != userID {
			return reviewdomain.ErrReviewNotFound
		}

		if err = s.articleRepo.DeleteArticleReview(ctx, reviewID, userID); err != nil {
			return fmt.Errorf("service.DeleteArticleReview: %w", err)
		}
		return nil
	})
}

// DeleteArticleReviewAdmin soft-deletes any article review, bypassing ownership checks.
func (s *Service) DeleteArticleReviewAdmin(ctx context.Context, reviewID, userID string) error {
	if err := s.articleRepo.DeleteArticleReview(ctx, reviewID, userID); err != nil {
		return fmt.Errorf("service.DeleteArticleReviewAdmin: %w", err)
	}
	return nil
}
