package courseservice

import (
	"context"
	"errors"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
)

// CreateCourse creates a new draft course after checking the slug is not already in use.
func (s *Service) CreateCourse(ctx context.Context, req coursedomain.CreateCourseRequest) (string, error) {
	existing, err := s.courseRepo.GetCourseBySlug(ctx, req.Slug)
	if err != nil && !errors.Is(err, coursedomain.ErrCourseNotFound) {
		return "", fmt.Errorf("service.CreateCourse: %w", err)
	}
	if existing != nil {
		return "", coursedomain.ErrInvalidSlug
	}

	isIndexable := true
	if req.IsIndexable != nil {
		isIndexable = *req.IsIndexable
	}

	course := coursedomain.Course{
		Slug:             req.Slug,
		Title:            req.Title,
		Description:      req.Description,
		ThumbnailURL:     req.ThumbnailURL,
		PreviewVideoURL:  req.PreviewVideoURL,
		EstimatedMinutes: req.EstimatedMinutes,
		SeoTitle:         req.SeoTitle,
		SeoDescription:   req.SeoDescription,
		OgImageURL:       req.OgImageURL,
		CanonicalURL:     req.CanonicalURL,
		IsIndexable:      isIndexable,
		CreatedByUserID:  req.CreatedByUserID,
	}

	createdCourse, err := s.courseRepo.CreateCourse(ctx, &course)
	if err != nil {
		return "", err
	}

	return createdCourse.ID, nil
}
