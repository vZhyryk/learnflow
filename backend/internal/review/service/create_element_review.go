package reviewservice

import (
	"context"
	"errors"
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
)

func (s *Service) CreateCourseReview(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		hasAccess, err := s.accessChecker.HasAccessCourse(ctx, req.UserID, req.CourseID)
		if err != nil {
			return fmt.Errorf("service.CreateCourseReview: check course access: %w", err)
		}

		if !hasAccess {
			return reviewdomain.ErrNoPermission
		}

		existingReview, err := s.courseRepo.GetCourseReviewByUserAndCourseID(ctx, req.UserID, req.CourseID)
		if err != nil && !errors.Is(err, reviewdomain.ErrReviewNotFound) {
			return fmt.Errorf("service.CreateCourseReview: check existing review: %w", err)
		}

		if existingReview != nil {
			return reviewdomain.ErrAlreadyReviewed
		}

		review := &reviewdomain.CourseReview{
			CourseID: req.CourseID,
			UserID:   req.UserID,
			Rating:   req.Rating,
			Comment:  req.Comment,
		}

		if _, err = s.courseRepo.CreateCourseReview(ctx, review); err != nil {
			return fmt.Errorf("service.CreateCourseReview: %w", err)
		}
		return nil
	})
}

func (s *Service) CreateContentReview(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		hasAccess, err := s.accessChecker.HasAccessContent(ctx, req.UserID, req.ContentID)
		if err != nil {
			return fmt.Errorf("service.CreateContentReview: check content access: %w", err)
		}

		if !hasAccess {
			return reviewdomain.ErrNoPermission
		}

		existingReview, err := s.contentRepo.GetContentReviewByUserAndContentID(ctx, req.UserID, req.ContentID)
		if err != nil && !errors.Is(err, reviewdomain.ErrReviewNotFound) {
			return fmt.Errorf("service.CreateContentReview: check existing review: %w", err)
		}

		if existingReview != nil {
			return reviewdomain.ErrAlreadyReviewed
		}

		review := &reviewdomain.ContentReview{
			ContentID: req.ContentID,
			UserID:    req.UserID,
			Rating:    req.Rating,
			Comment:   req.Comment,
		}

		if _, err := s.contentRepo.CreateContentReview(ctx, review); err != nil {
			return fmt.Errorf("service.CreateContentReview: %w", err)
		}
		return nil
	})
}

func (s *Service) CreateCourseReviewAdmin(ctx context.Context, req reviewdomain.CreateCourseReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		existingReview, err := s.courseRepo.GetCourseReviewByUserAndCourseID(ctx, req.UserID, req.CourseID)
		if err != nil && !errors.Is(err, reviewdomain.ErrReviewNotFound) {
			return fmt.Errorf("service.CreateCourseReviewAdmin: check existing review: %w", err)
		}
		if existingReview != nil {
			return reviewdomain.ErrAlreadyReviewed
		}

		review := &reviewdomain.CourseReview{
			CourseID: req.CourseID,
			UserID:   req.UserID,
			Rating:   req.Rating,
			Comment:  req.Comment,
		}
		if _, err := s.courseRepo.CreateCourseReview(ctx, review); err != nil {
			return fmt.Errorf("service.CreateCourseReviewAdmin: %w", err)
		}
		return nil
	})
}

func (s *Service) CreateContentReviewAdmin(ctx context.Context, req reviewdomain.CreateContentReviewRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		existingReview, err := s.contentRepo.GetContentReviewByUserAndContentID(ctx, req.UserID, req.ContentID)
		if err != nil && !errors.Is(err, reviewdomain.ErrReviewNotFound) {
			return fmt.Errorf("service.CreateContentReviewAdmin: check existing review: %w", err)
		}
		if existingReview != nil {
			return reviewdomain.ErrAlreadyReviewed
		}

		review := &reviewdomain.ContentReview{
			ContentID: req.ContentID,
			UserID:    req.UserID,
			Rating:    req.Rating,
			Comment:   req.Comment,
		}

		if _, err := s.contentRepo.CreateContentReview(ctx, review); err != nil {
			return fmt.Errorf("service.CreateContentReviewAdmin: %w", err)
		}
		return nil
	})
}
