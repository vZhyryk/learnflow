package reviewservice

import (
	"context"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
)

// UpdateCourseReview updates a course review owned by the requesting user.
func (s *Service) UpdateCourseReview(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.courseRepo.GetCourseReviewByID(ctx, req.ReviewID)
		if err != nil {
			return fmt.Errorf("service.UpdateCourseReview: fetch review: %w", err)
		}

		if currentReview.UserID != req.UserID {
			return reviewdomain.ErrReviewNotFound
		}

		hasAccess, err := s.accessChecker.HasAccessCourse(ctx, req.UserID, currentReview.CourseID)
		if err != nil {
			return fmt.Errorf("service.UpdateCourseReview: check course access: %w", err)
		}

		if !hasAccess {
			return reviewdomain.ErrNoPermission
		}

		req.Apply(currentReview)

		if err = s.courseRepo.UpdateCourseReview(ctx, currentReview); err != nil {
			return fmt.Errorf("service.UpdateCourseReview: %w", err)
		}
		return nil
	})
}

// UpdateContentReview updates a content review owned by the requesting user.
func (s *Service) UpdateContentReview(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.contentRepo.GetContentReviewByID(ctx, req.ReviewID)
		if err != nil {
			return fmt.Errorf("service.UpdateContentReview: fetch review: %w", err)
		}

		if currentReview.UserID != req.UserID {
			return reviewdomain.ErrReviewNotFound
		}

		hasAccess, err := s.accessChecker.HasAccessContent(ctx, req.UserID, currentReview.ContentID)
		if err != nil {
			return fmt.Errorf("service.UpdateContentReview: check content access: %w", err)
		}

		if !hasAccess {
			return reviewdomain.ErrNoPermission
		}

		req.Apply(currentReview)

		if err := s.contentRepo.UpdateContentReview(ctx, currentReview); err != nil {
			return fmt.Errorf("service.UpdateContentReview: %w", err)
		}
		return nil
	})
}

// UpdateCourseReviewAdmin updates any course review, bypassing ownership/access checks.
func (s *Service) UpdateCourseReviewAdmin(ctx context.Context, req reviewdomain.UpdateCourseReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.courseRepo.GetCourseReviewByID(ctx, req.ReviewID)
		if err != nil {
			return fmt.Errorf("service.UpdateCourseReviewAdmin: fetch review: %w", err)
		}

		req.Apply(currentReview)

		if err := s.courseRepo.UpdateCourseReview(ctx, currentReview); err != nil {
			return fmt.Errorf("service.UpdateCourseReviewAdmin: %w", err)
		}
		return nil
	})
}

// UpdateContentReviewAdmin updates any content review, bypassing ownership/access checks.
func (s *Service) UpdateContentReviewAdmin(ctx context.Context, req reviewdomain.UpdateContentReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.contentRepo.GetContentReviewByID(ctx, req.ReviewID)
		if err != nil {
			return fmt.Errorf("service.UpdateContentReviewAdmin: fetch review: %w", err)
		}

		req.Apply(currentReview)

		if err := s.contentRepo.UpdateContentReview(ctx, currentReview); err != nil {
			return fmt.Errorf("service.UpdateContentReviewAdmin: %w", err)
		}
		return nil
	})
}

// UpdateArticleReview updates an article review owned by the requesting user.
func (s *Service) UpdateArticleReview(ctx context.Context, req reviewdomain.UpdateArticleReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.articleRepo.GetArticleReviewByID(ctx, req.ReviewID)
		if err != nil {
			return fmt.Errorf("service.UpdateArticleReview: fetch review: %w", err)
		}

		if currentReview.UserID != req.UserID {
			return reviewdomain.ErrReviewNotFound
		}

		req.Apply(currentReview)

		if err = s.articleRepo.UpdateArticleReview(ctx, currentReview); err != nil {
			return fmt.Errorf("service.UpdateArticleReview: %w", err)
		}
		return nil
	})
}

// UpdateArticleReviewAdmin updates any article review, bypassing ownership checks.
func (s *Service) UpdateArticleReviewAdmin(ctx context.Context, req reviewdomain.UpdateArticleReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		currentReview, err := s.articleRepo.GetArticleReviewByID(ctx, req.ReviewID)
		if err != nil {
			return fmt.Errorf("service.UpdateArticleReviewAdmin: fetch review: %w", err)
		}

		req.Apply(currentReview)

		if err := s.articleRepo.UpdateArticleReview(ctx, currentReview); err != nil {
			return fmt.Errorf("service.UpdateArticleReviewAdmin: %w", err)
		}
		return nil
	})
}
